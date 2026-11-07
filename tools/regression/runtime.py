"""Run only newly created, isolated application data directories."""

import os
import secrets
import socket
import subprocess
import time
from pathlib import Path

from client import Client


class Runtime:
    def __init__(self, binary, backend, static_dir, data_dir, log_dir, image=None, windows_default=False,
                 configuration=None):
        self.binary = Path(binary).resolve()
        self.backend = Path(backend).resolve()
        self.static_dir = Path(static_dir).resolve()
        self.data_dir = Path(data_dir).resolve()
        self.log_dir = Path(log_dir).resolve()
        self.process = None
        self.log = None
        self.client = None
        self.image = image
        self.container = None
        self.windows_default = windows_default
        self.configuration = configuration or {}
        if set(self.configuration) - {"RAILKEEPER_MAX_IMAGE_MB", "RAILKEEPER_MAX_ATTACHMENT_MB"}:
            raise ValueError("Only bounded upload limits are accepted in the synthetic launcher configuration")

    def __enter__(self):
        if self.windows_default:
            self.data_dir.parents[1].mkdir(mode=0o700, parents=True, exist_ok=True)
        else:
            self.data_dir.mkdir(mode=0o700, parents=True, exist_ok=True)
        self.log_dir.mkdir(mode=0o700, parents=True, exist_ok=True)
        with socket.socket() as probe:
            probe.bind(("127.0.0.1", 0))
            port = probe.getsockname()[1]
        env = {key: value for key, value in os.environ.items() if not key.startswith("RAILKEEPER_")}
        env.update({
            "RAILKEEPER_ADDR": f"127.0.0.1:{port}", "RAILKEEPER_DATA_DIR": str(self.data_dir),
            "RAILKEEPER_MIGRATIONS_DIR": str(self.backend / "migrations"),
            "RAILKEEPER_SEEDS_DIR": str(self.backend / "seeds"),
            "RAILKEEPER_STATIC_DIR": str(self.static_dir), "RAILKEEPER_OPEN_BROWSER": "false",
            "RAILKEEPER_UPDATE_CHECK_URL": f"http://127.0.0.1:{port}/unavailable-update-feed",
        })
        env.update(self.configuration)
        self.log = (self.log_dir / "server.log").open("wb")
        command = [str(self.binary)]
        if self.windows_default:
            if os.name != "nt" or self.data_dir.parts[-2:] != ("RailKeeper", "data"):
                raise ValueError("Windows default-path test requires an isolated LOCALAPPDATA/RailKeeper/data")
            env.pop("RAILKEEPER_DATA_DIR")
            env["LOCALAPPDATA"] = str(self.data_dir.parents[1])
            command.append("--standalone")
        if self.image:
            self.container = "railkeeper-regression-" + secrets.token_hex(8)
            command = ["docker", "run", "--rm", "--name", self.container, "--read-only",
                       "--cap-drop", "ALL", "--security-opt", "no-new-privileges:true", "--tmpfs", "/tmp",
                       "--user", "0:0", "--publish", f"127.0.0.1:{port}:8080",
                       "--mount", f"type=bind,source={self.data_dir},target=/data",
                       "--env", "RAILKEEPER_ADDR=:8080", "--env", "RAILKEEPER_DATA_DIR=/data",
                       "--env", "RAILKEEPER_UPDATE_CHECK_URL=http://127.0.0.1:8080/unavailable-update-feed"]
            for key, value in self.configuration.items():
                command.extend(["--env", f"{key}={value}"])
            command.append(self.image)
        self.process = subprocess.Popen(command, cwd=self.log_dir, env=env,
                                        stdout=self.log, stderr=subprocess.STDOUT)
        self.client = Client(port)
        for _ in range(200):
            if self.process.poll() is not None:
                self.close()
                raise AssertionError("Regression server exited; inspect its local server.log")
            try:
                self.client.request("GET", "/health")
                code_file = self.data_dir / "setup-code.txt"
                if code_file.is_file():
                    self.client.setup_code = code_file.read_text(encoding="utf-8").strip()
                return self
            except OSError:
                time.sleep(0.1)
        self.close()
        raise AssertionError("Regression server did not become healthy in 20 seconds")

    def crash(self):
        if self.container or self.process is None or self.process.poll() is not None:
            raise ValueError("Only this runner's live native process can be crash-tested")
        self.process.kill()
        self.process.wait(timeout=10)

    def close(self):
        if self.container:
            # Only this invocation's randomly named container is removed.
            subprocess.run(["docker", "rm", "--force", "--volumes", self.container],
                           stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL, timeout=30)
            self.container = None
        if self.process is not None and self.process.poll() is None:
            self.process.terminate()
            try:
                self.process.wait(timeout=10)
            except subprocess.TimeoutExpired:
                self.process.kill()
                self.process.wait(timeout=10)
        if self.log is not None:
            self.log.close()

    def __exit__(self, *_):
        self.close()
