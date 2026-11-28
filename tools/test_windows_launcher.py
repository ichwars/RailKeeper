"""Exercise the relocated, extracted Windows ZIP on the native Windows runner."""

from contextlib import closing
import ctypes as ct
from ctypes import wintypes as win
import os
from pathlib import Path
import socket
import sqlite3
import subprocess
import sys
import tempfile
import time
import urllib.error
import urllib.request


def api(dll, name, result, *args):
    function = getattr(dll, name)
    function.restype = result
    function.argtypes = args
    return function


kernel = ct.WinDLL("kernel32", use_last_error=True)
user = ct.WinDLL("user32", use_last_error=True)
attach_console = api(kernel, "AttachConsole", win.BOOL, win.DWORD)
free_console = api(kernel, "FreeConsole", win.BOOL)
console_window = api(kernel, "GetConsoleWindow", win.HWND)
console_processes = api(kernel, "GetConsoleProcessList", win.DWORD, ct.POINTER(win.DWORD), win.DWORD)
ignore_ctrl_c = api(kernel, "SetConsoleCtrlHandler", win.BOOL, ct.c_void_p, win.BOOL)
send_ctrl_event = api(kernel, "GenerateConsoleCtrlEvent", win.BOOL, win.DWORD, win.DWORD)
open_process = api(kernel, "OpenProcess", win.HANDLE, win.DWORD, win.BOOL, win.DWORD)
process_path = api(kernel, "QueryFullProcessImageNameW", win.BOOL,
                   win.HANDLE, win.DWORD, win.LPWSTR, ct.POINTER(win.DWORD))
wait_for_process = api(kernel, "WaitForSingleObject", win.DWORD, win.HANDLE, win.DWORD)
get_exit_code = api(kernel, "GetExitCodeProcess", win.BOOL, win.HANDLE, ct.POINTER(win.DWORD))
terminate_process = api(kernel, "TerminateProcess", win.BOOL, win.HANDLE, win.UINT)
close_handle = api(kernel, "CloseHandle", win.BOOL, win.HANDLE)
enum_windows_callback = ct.WINFUNCTYPE(win.BOOL, win.HWND, win.LPARAM)
enum_windows = api(user, "EnumWindows", win.BOOL, enum_windows_callback, win.LPARAM)
window_pid = api(user, "GetWindowThreadProcessId", win.DWORD, win.HWND, ct.POINTER(win.DWORD))
get_window_class = api(user, "GetClassNameW", ct.c_int, win.HWND, win.LPWSTR, ct.c_int)
get_window_text = api(user, "GetWindowTextW", ct.c_int, win.HWND, win.LPWSTR, ct.c_int)
is_visible = api(user, "IsWindowVisible", win.BOOL, win.HWND)
post_message = api(user, "PostMessageW", win.BOOL, win.HWND, win.UINT, win.WPARAM, win.LPARAM)
is_window = api(user, "IsWindow", win.BOOL, win.HWND)
is_minimized = api(user, "IsIconic", win.BOOL, win.HWND)

PROCESS_TERMINATE = 0x0001
PROCESS_QUERY_LIMITED_INFORMATION = 0x1000
SYNCHRONIZE = 0x00100000
WAIT_OBJECT_0 = 0
WAIT_TIMEOUT = 258
CTRL_C_EVENT = 0
WM_CLOSE = 0x0010


def require(ok, message):
    if not ok:
        raise RuntimeError(message)


def require_win32(ok, operation):
    if not ok:
        raise OSError(f"{operation}: {ct.WinError(ct.get_last_error())}")


def wait_until(predicate, message, seconds=30):
    deadline = time.monotonic() + seconds
    while not predicate():
        require(time.monotonic() < deadline, message)
        time.sleep(0.1)


def find_desktop_window(pid):
    found = None

    @enum_windows_callback
    def find(hwnd, _):
        nonlocal found
        owner = win.DWORD()
        window_pid(hwnd, ct.byref(owner))
        if owner.value == pid and is_visible(hwnd):
            name = ct.create_unicode_buffer(64)
            title = ct.create_unicode_buffer(64)
            get_window_class(hwnd, name, len(name))
            get_window_text(hwnd, title, len(title))
            # The native launcher sets this title after successful WebView2 navigation.
            if name.value == "RailKeeperDesktopWindow" and title.value == "RailKeeper":
                found = hwnd
        return True

    require_win32(enum_windows(find, 0), "Find the WebView2 desktop window")
    return found


def assert_launcher_lifecycle(package, address, working_dir, *, shutdown):
    launcher = subprocess.Popen([str(package / "Start RailKeeper.exe")], cwd=working_dir)
    try:
        assert_server_lifecycle(package, address, launcher, shutdown=shutdown)
    finally:
        if launcher.poll() is None:
            launcher.kill()
        launcher.wait(timeout=10)

    database = working_dir / "railkeeper.db"
    with closing(sqlite3.connect(f"{database.as_uri()}?mode=rw", uri=True)) as reopened:
        require(reopened.execute("PRAGMA quick_check").fetchone() == ("ok",),
                "SQLite could not be reopened cleanly after shutdown.")


def assert_server_lifecycle(package, address, launcher, *, shutdown):
    opener = urllib.request.build_opener(urllib.request.ProxyHandler({}))

    def ready():
        try:
            with opener.open(f"http://{address}/health", timeout=1) as response:
                return response.status == 200
        except (urllib.error.URLError, TimeoutError):
            return False

    wait_until(ready, "RailKeeper did not become ready from the relocated launcher.")
    wait_until(lambda: find_desktop_window(launcher.pid),
               "RailKeeper did not finish loading its local app in a visible WebView2 window.", seconds=60)
    desktop = find_desktop_window(launcher.pid)
    port = int(address.rsplit(":", 1)[1])
    require(launcher.poll() is None, "The launcher did not stay alive to supervise its server.")
    # Find the process serving this test's isolated endpoint.
    pid = int(subprocess.check_output([
        "powershell.exe", "-NoProfile", "-NonInteractive", "-Command",
        f"(Get-NetTCPConnection -LocalAddress 127.0.0.1 -LocalPort {port} -State Listen -ErrorAction Stop).OwningProcess",
    ], text=True, timeout=20, creationflags=subprocess.CREATE_NO_WINDOW).strip())
    process = open_process(SYNCHRONIZE | PROCESS_QUERY_LIMITED_INFORMATION | PROCESS_TERMINATE, False, pid)
    require_win32(process, "Open the RailKeeper process serving the test endpoint")
    verified_process = False
    try:
        path = ct.create_unicode_buffer(32768)
        size = win.DWORD(len(path))
        require_win32(process_path(process, 0, path, ct.byref(size)), "Read the RailKeeper process path")
        require(Path(path.value).samefile(package / "RailKeeper.exe"),
                f"Starter launched an executable outside the extracted ZIP: {path.value}")

        verified_process = True
        require_win32(attach_console(pid), "Attach to the RailKeeper console")
        try:
            window = console_window()
            require(window and is_window(window), "RailKeeper has no console window.")
            require(is_minimized(window), "The starter opened a console that was not minimized.")
            print("The relocated starter launched RailKeeper with a minimized console window.")

            if shutdown == "pipe":
                # Only terminate the test's launcher. Windows closes its private
                # shutdown pipe, which must make the server exit on its own.
                launcher.terminate()
                launcher.wait(timeout=10)
            elif shutdown == "window X":
                require_win32(post_message(desktop, WM_CLOSE, 0, 0), "Close the RailKeeper desktop window")
            else:
                send_console_interrupt(pid)

            require(wait_for_process(process, 15000) == WAIT_OBJECT_0, "RailKeeper did not exit after shutdown.")
            exit_code = win.DWORD()
            require_win32(get_exit_code(process, ct.byref(exit_code)), "Read the RailKeeper exit code")
            require(exit_code.value == 0, f"Unexpected shutdown exit code: {exit_code.value:#x}")
        finally:
            require_win32(free_console(), "Detach the probe from the RailKeeper console")

        wait_until(lambda: not is_window(window), "The console window remained after RailKeeper exited.", seconds=10)
        wait_until(lambda: not is_window(desktop), "The desktop window remained after RailKeeper exited.", seconds=10)
        wait_until(lambda: not ready(), "RailKeeper's listener remained after shutdown.", seconds=5)
        if shutdown == "pipe":
            print("Closing the launcher's pipe stopped RailKeeper cleanly and closed its console.")
        else:
            require(launcher.wait(timeout=10) == 0, "The launcher did not report a clean exit.")
            print(f"{shutdown} stopped RailKeeper cleanly and closed the WebView2 window and console.")
    finally:
        if verified_process and wait_for_process(process, 0) == WAIT_TIMEOUT:
            terminate_process(process, 1)
            wait_for_process(process, 5000)
        close_handle(process)


def send_console_interrupt(pid):
    attached = (win.DWORD * 8)()
    count = console_processes(attached, len(attached))
    require(count == 2 and set(attached[:count]) == {pid, os.getpid()},
            "Refusing to send Ctrl+C to a console shared with other processes.")
    require_win32(ignore_ctrl_c(None, True), "Keep the console probe alive during Ctrl+C")
    require_win32(send_ctrl_event(CTRL_C_EVENT, 0), "Send Ctrl+C to the RailKeeper console")


def assert_batch_failure_returns_to_prompt(package, working_dir):
    env = dict(os.environ, RAILKEEPER_ADDR="127.0.0.1:not-a-port")
    commands = (
        f'call "{package / "start-railkeeper.bat"}"\n'
        "echo RAILKEEPER_TEST_EXIT=!errorlevel!\n"
        "echo RAILKEEPER_TEST_CWD=!cd!\n"
        "exit /b 0\n"
    )
    result = subprocess.run(
        [os.environ["COMSPEC"], "/d", "/q", "/v:on"], input=commands, text=True,
        stdout=subprocess.PIPE, stderr=subprocess.STDOUT, errors="replace",
        creationflags=subprocess.CREATE_NEW_CONSOLE, cwd=working_dir, env=env, timeout=20, check=True,
    )
    # With redirected input, cmd writes its prompt and the echo result on the same line.
    lines = [line.rpartition(">")[2] for line in result.stdout.splitlines()]
    require("server listen failed" in result.stdout and "RAILKEEPER_TEST_EXIT=1" in lines,
            f"The batch launcher did not propagate the application startup failure:\n{result.stdout}")
    marker = "RAILKEEPER_TEST_CWD="
    cwd = next((Path(line.removeprefix(marker)) for line in lines if line.startswith(marker)), None)
    # cmd may display the short Windows path even when the caller used a long path.
    require(cwd is not None and cwd.samefile(working_dir),
            f"The caller's Command Prompt did not retain its working directory:\n{result.stdout}")
    print("The batch launcher returned exit code 1 and restored the calling Command Prompt's directory.")


def main():
    package = Path(sys.argv[1]).resolve(strict=True)
    require_win32(free_console(), "Detach from the test runner console")
    with tempfile.TemporaryDirectory(prefix="railkeeper-launcher-data-") as root:
        for shutdown in ("window X", "Ctrl+C", "pipe"):
            data = Path(root) / shutdown
            data.mkdir()
            with socket.socket() as listener:
                listener.bind(("127.0.0.1", 0))
                address = f"127.0.0.1:{listener.getsockname()[1]}"
            os.environ.update(RAILKEEPER_ADDR=address, RAILKEEPER_DATA_DIR=str(data))
            assert_launcher_lifecycle(package, address, data, shutdown=shutdown)
        assert_batch_failure_returns_to_prompt(package, Path(root))


if __name__ == "__main__":
    main()
