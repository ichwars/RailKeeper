"""HTTP client restricted to a regression process's allocated loopback port."""

import http.cookiejar
import json
import secrets
import urllib.error
import urllib.request


class Client:
    def __init__(self, port):
        if not isinstance(port, int) or not 0 < port < 65536:
            raise ValueError("A local regression port is required")
        self.base = f"http://127.0.0.1:{port}"
        self.token = ""
        self.cookies = http.cookiejar.CookieJar()
        self.opener = urllib.request.build_opener(
            urllib.request.ProxyHandler({}), urllib.request.HTTPCookieProcessor(self.cookies)
        )

    def request(self, method, path, data=None, expected=200, content_type="application/json", csrf=True, raw=False):
        if not path.startswith("/api/v1/") and path != "/health":
            raise ValueError("Only local application endpoints are allowed")
        payload = json.dumps(data).encode() if data is not None and not isinstance(data, bytes) else data
        headers = {"Content-Type": content_type, "Origin": self.base}
        if csrf and self.token:
            headers["X-CSRF-Token"] = self.token
        request = urllib.request.Request(self.base + path, data=payload, method=method, headers=headers)
        try:
            response = self.opener.open(request, timeout=60)
        except urllib.error.HTTPError as error:
            response = error
        with response:
            body = response.read()
            if response.status != expected:
                raise AssertionError(f"{method} {path}: HTTP {response.status}, expected {expected}: {body[:300]!r}")
            if not raw and "json" in response.headers.get("Content-Type", "") and body:
                return json.loads(body)
            return body

    def login(self, credentials):
        result = self.request("POST", "/api/v1/auth/login", credentials)
        self.token = result["csrfToken"]
        return result

    def upload(self, path, name, content, mime="application/octet-stream", fields=None, expected=201):
        boundary = "railkeeper-regression-" + secrets.token_hex(12)
        parts = []
        for key, value in (fields or {}).items():
            parts.append(f'--{boundary}\r\nContent-Disposition: form-data; name="{key}"\r\n\r\n{value}\r\n'.encode())
        parts.extend([
            f'--{boundary}\r\nContent-Disposition: form-data; name="file"; filename="{name}"\r\nContent-Type: {mime}\r\n\r\n'.encode(),
            content, f"\r\n--{boundary}--\r\n".encode(),
        ])
        return self.request("POST", path, b"".join(parts), expected,
                            content_type=f"multipart/form-data; boundary={boundary}")


def setup(client, username):
    credentials = {"username": username, "password": secrets.token_urlsafe(32)}
    client.request("POST", "/api/v1/setup/admin", {**credentials, "email": username + "@example.test"}, 201)
    client.login(credentials)
    return credentials
