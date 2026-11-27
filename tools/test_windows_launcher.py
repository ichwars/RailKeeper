"""Exercise the relocated, extracted Windows ZIP using the Windows shell."""

import ctypes as ct
from ctypes import wintypes as win
import os
from pathlib import Path
import socket
import subprocess
import sys
import tempfile
import time
import urllib.error
import urllib.request


class ShellExecuteInfo(ct.Structure):
    _fields_ = [
        ("cbSize", win.DWORD), ("fMask", win.ULONG), ("hwnd", win.HWND),
        ("lpVerb", win.LPCWSTR), ("lpFile", win.LPCWSTR),
        ("lpParameters", win.LPCWSTR), ("lpDirectory", win.LPCWSTR),
        ("nShow", ct.c_int), ("hInstApp", win.HINSTANCE), ("lpIDList", ct.c_void_p),
        ("lpClass", win.LPCWSTR), ("hkeyClass", win.HKEY), ("dwHotKey", win.DWORD),
        ("hIcon", win.HANDLE), ("hProcess", win.HANDLE),
    ]


def api(dll, name, result, *args):
    function = getattr(dll, name)
    function.restype = result
    function.argtypes = args
    return function


kernel = ct.WinDLL("kernel32", use_last_error=True)
user = ct.WinDLL("user32", use_last_error=True)
shell = ct.WinDLL("shell32", use_last_error=True)
ole = ct.OleDLL("ole32")

shell_execute = api(shell, "ShellExecuteExW", win.BOOL, ct.POINTER(ShellExecuteInfo))
attach_console = api(kernel, "AttachConsole", win.BOOL, win.DWORD)
free_console = api(kernel, "FreeConsole", win.BOOL)
console_window = api(kernel, "GetConsoleWindow", win.HWND)
console_processes = api(kernel, "GetConsoleProcessList", win.DWORD, ct.POINTER(win.DWORD), win.DWORD)
ignore_ctrl_c = api(kernel, "SetConsoleCtrlHandler", win.BOOL, ct.c_void_p, win.BOOL)
send_ctrl_event = api(kernel, "GenerateConsoleCtrlEvent", win.BOOL, win.DWORD, win.DWORD)
process_id = api(kernel, "GetProcessId", win.DWORD, win.HANDLE)
process_path = api(kernel, "QueryFullProcessImageNameW", win.BOOL,
                   win.HANDLE, win.DWORD, win.LPWSTR, ct.POINTER(win.DWORD))
wait_for_process = api(kernel, "WaitForSingleObject", win.DWORD, win.HANDLE, win.DWORD)
get_exit_code = api(kernel, "GetExitCodeProcess", win.BOOL, win.HANDLE, ct.POINTER(win.DWORD))
terminate_process = api(kernel, "TerminateProcess", win.BOOL, win.HANDLE, win.UINT)
close_handle = api(kernel, "CloseHandle", win.BOOL, win.HANDLE)
is_window = api(user, "IsWindow", win.BOOL, win.HWND)
is_minimized = api(user, "IsIconic", win.BOOL, win.HWND)

WAIT_OBJECT_0 = 0
WAIT_TIMEOUT = 258
CTRL_C_EVENT = 0
STATUS_CONTROL_C_EXIT = 0xC000013A


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


def open_shortcut(path):
    ole.CoInitializeEx(None, 2)  # COINIT_APARTMENTTHREADED
    try:
        info = ShellExecuteInfo()
        info.cbSize = ct.sizeof(info)
        # Keep the process handle, wait for launch, and suppress failure dialogs.
        info.fMask = 0x40 | 0x100 | 0x400
        info.lpFile = str(path)
        info.lpDirectory = str(path.parent)
        info.nShow = 1  # Normal shell invocation; the shortcut supplies its own show mode.
        require_win32(shell_execute(ct.byref(info)), "Open the extracted Start RailKeeper shortcut")
        require(info.hProcess, "The shortcut returned no process handle.")
        return info.hProcess
    finally:
        ole.CoUninitialize()


def assert_shortcut_lifecycle(package, address):
    # The shell launch and its Ctrl+C must never share the CI runner's console.
    require_win32(free_console(), "Detach from the test runner console")
    process = open_shortcut(package / "Start RailKeeper.lnk")
    try:
        pid = process_id(process)
        require_win32(pid, "Identify the shortcut process")
        path = ct.create_unicode_buffer(32768)
        size = win.DWORD(len(path))
        require_win32(process_path(process, 0, path, ct.byref(size)), "Read the shortcut process path")
        require(Path(path.value).samefile(package / "RailKeeper.exe"),
                f"Shortcut launched an executable outside the extracted ZIP: {path.value}")

        opener = urllib.request.build_opener(urllib.request.ProxyHandler({}))

        def ready():
            require(wait_for_process(process, 0) == WAIT_TIMEOUT, "RailKeeper exited during startup.")
            try:
                with opener.open(f"http://{address}/health", timeout=1) as response:
                    return response.status == 200
            except (urllib.error.URLError, TimeoutError):
                return False

        wait_until(ready, "RailKeeper did not become ready from the relocated shortcut.")
        require_win32(attach_console(pid), "Attach to the RailKeeper console")
        try:
            window = console_window()
            require(window and is_window(window), "RailKeeper has no console window.")
            require(is_minimized(window), "The shortcut opened a console that was not minimized.")
            print("The relocated shortcut launched RailKeeper with a minimized console window.")

            attached = (win.DWORD * 8)()
            count = console_processes(attached, len(attached))
            require(count == 2 and set(attached[:count]) == {pid, os.getpid()},
                    "Refusing to send Ctrl+C to a console shared with other processes.")
            require_win32(ignore_ctrl_c(None, True), "Keep the console probe alive during Ctrl+C")
            require_win32(send_ctrl_event(CTRL_C_EVENT, 0), "Send Ctrl+C to the RailKeeper console")
            require(wait_for_process(process, 15000) == WAIT_OBJECT_0, "RailKeeper did not exit after Ctrl+C.")
            exit_code = win.DWORD()
            require_win32(get_exit_code(process, ct.byref(exit_code)), "Read the RailKeeper exit code")
            require(exit_code.value in {0, STATUS_CONTROL_C_EXIT},
                    f"Unexpected exit code after Ctrl+C: {exit_code.value:#x}")
        finally:
            require_win32(free_console(), "Detach the probe from the RailKeeper console")

        wait_until(lambda: not is_window(window), "The console window remained after RailKeeper exited.", seconds=10)
        print("Ctrl+C stopped RailKeeper, and the console window closed after the probe detached.")
    finally:
        if wait_for_process(process, 0) == WAIT_TIMEOUT:
            terminate_process(process, 1)
            wait_for_process(process, 5000)
        close_handle(process)


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
    lines = result.stdout.splitlines()
    require("server listen failed" in result.stdout and "RAILKEEPER_TEST_EXIT=1" in lines,
            f"The batch launcher did not propagate the application startup failure:\n{result.stdout}")
    require(f"RAILKEEPER_TEST_CWD={working_dir}".casefold() in {line.casefold() for line in lines},
            f"The caller's Command Prompt did not retain its working directory:\n{result.stdout}")
    print("The batch launcher returned exit code 1 and restored the calling Command Prompt's directory.")


def main():
    package = Path(sys.argv[1]).resolve(strict=True)
    with tempfile.TemporaryDirectory(prefix="railkeeper-launcher-data-") as data:
        with socket.socket() as listener:
            listener.bind(("127.0.0.1", 0))
            address = f"127.0.0.1:{listener.getsockname()[1]}"
        os.environ.update(RAILKEEPER_ADDR=address, RAILKEEPER_DATA_DIR=data, RAILKEEPER_OPEN_BROWSER="false")
        assert_shortcut_lifecycle(package, address)
        assert_batch_failure_returns_to_prompt(package, Path(data))


if __name__ == "__main__":
    main()
