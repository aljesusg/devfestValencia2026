#!/usr/bin/env python3
"""Launch all four workshop profiles in one MCP Inspector."""

import argparse
import json
import os
from pathlib import Path
import socket
import subprocess
import tempfile
import time
import urllib.request


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--client-port", type=int, default=6294)
    parser.add_argument("--http-port", type=int, default=8093)
    parser.add_argument("--check", action="store_true", help="build and check HTTP without launching Inspector")
    args = parser.parse_args()
    root = Path(__file__).resolve().parent
    ports = [args.http_port] if args.check else [
        args.http_port, args.client_port, args.client_port + 1,
        args.client_port + 3, args.client_port + 4,
    ]
    if len(set(ports)) != len(ports):
        parser.error("HTTP and Inspector ports must not overlap")
    for port in ports:
        if not 1 <= port <= 65535:
            parser.error(f"invalid port: {port}")
        with socket.socket() as sock:
            sock.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
            try:
                sock.bind(("127.0.0.1", port))
            except OSError:
                parser.error(f"port {port} is busy; choose --client-port or --http-port")

    with tempfile.TemporaryDirectory(prefix="devfest-inspector-") as temp:
        binaries = {}
        for folder in ("01-hello-stdio", "02-tools-schemas", "03-http-transport"):
            binary = str(Path(temp) / folder)
            subprocess.run(["go", "build", "-o", binary, "."], cwd=root / folder, check=True)
            binaries[folder] = binary
        url = f"http://127.0.0.1:{args.http_port}"
        env = {**os.environ, "SCHEDULE_PATH": str(root / "data/schedule.json")}
        # Avoid an inherited PORT overriding the explicit local HTTP address.
        env.pop("PORT", None)
        backend = subprocess.Popen(
            [binaries["03-http-transport"], f"-http=127.0.0.1:{args.http_port}"], env=env,
        )
        try:
            for _ in range(100):
                if backend.poll() is not None:
                    raise RuntimeError("HTTP server exited before becoming ready")
                try:
                    with urllib.request.urlopen(url + "/health", timeout=0.5) as response:
                        assert json.load(response)["status"] == "ok"
                    break
                except OSError:
                    time.sleep(0.1)
            else:
                raise RuntimeError("HTTP server did not become ready")
            profiles = {
                "01 — Hello MCP · Valencian greeting": {
                    "command": binaries["01-hello-stdio"],
                },
                "02 — Tools & schemas · Explore the dataset": {
                    "command": binaries["02-tools-schemas"],
                    "env": {"SCHEDULE_PATH": env["SCHEDULE_PATH"]},
                },
                "03 — HTTP transport · Agenda analytics": {"type": "streamable-http", "url": url + "/mcp"},
                "04 — Connect a client · Guided investigation": {"type": "streamable-http", "url": url + "/mcp"},
            }
            config = Path(temp) / "inspector.json"
            config.write_text(json.dumps({"mcpServers": profiles}, indent=2))
            print(f"\nHTTP server: {url}/mcp\nGuide: {root / 'INSPECTOR.md'}", flush=True)
            if args.check:
                print("All three servers built; HTTP health passed; four profiles generated.")
                return
            inspector_env = {
                **os.environ, "CLIENT_PORT": str(args.client_port),
                "SERVER_PORT": str(args.client_port + 3),
                "MCP_SANDBOX_PORT": str(args.client_port + 1),
                "MCP_APP_ORIGIN_PORT": str(args.client_port + 4),
            }
            subprocess.run(
                ["npx", "@modelcontextprotocol/inspector", "--config", str(config)],
                env=inspector_env, check=True,
            )
        finally:
            backend.terminate()
            try:
                backend.wait(timeout=5)
            except subprocess.TimeoutExpired:
                backend.kill()
                backend.wait()


if __name__ == "__main__":
    try:
        main()
    except KeyboardInterrupt:
        pass
