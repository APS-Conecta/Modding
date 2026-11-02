#!/usr/bin/env python3
"""Exercise the real wizard with synthetic establishments; never provisions AIO."""
import argparse
import json
import re
import socket
import subprocess
import tempfile
import time
from pathlib import Path

from playwright.sync_api import expect, sync_playwright


def launch(binary, directory, fixture, secure=False):
    with socket.socket() as sock:
        sock.bind(("127.0.0.1", 0))
        port = sock.getsockname()[1]
    log_path = directory / "server.log"
    tls_args = []
    if secure:
        cert, key = directory / "test.crt", directory / "test.key"
        subprocess.run(["openssl", "req", "-x509", "-newkey", "rsa:2048", "-nodes",
                        "-keyout", str(key), "-out", str(cert), "-days", "1",
                        "-subj", "/CN=127.0.0.1", "-addext", "subjectAltName=IP:127.0.0.1"],
                       check=True, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
        tls_args = ["--tls-cert", str(cert), "--tls-key", str(key)]
    with log_path.open("w") as log:
        process = subprocess.Popen(
            [str(binary), "--listen", f"127.0.0.1:{port}", "--state-dir", str(directory / "state"),
             "--catalog", str(fixture), "--catalog-version", "synthetic-smoke-v1"] + tls_args,
            stdout=subprocess.DEVNULL, stderr=log,
        )
    deadline = time.monotonic() + 15
    while time.monotonic() < deadline:
        text = log_path.read_text()
        match = re.search(r"Código de acceso de un solo uso: ([a-f0-9]{64})", text)
        if match:
            scheme = "https" if secure else "http"
            return process, f"{scheme}://127.0.0.1:{port}", match.group(1)
        if process.poll() is not None:
            raise RuntimeError("Wizard exited before becoming ready; inspect private server log.")
        time.sleep(0.05)
    process.terminate()
    process.wait(timeout=5)
    raise RuntimeError("Wizard did not become ready.")


def login(page, url, token):
    page.goto(url)
    page.get_by_label("Código de acceso", exact=True).fill(token)
    page.get_by_role("button", name="Entrar al asistente").click()
    expect(page.get_by_role("heading", name="Elija su establecimiento")).to_be_visible()


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--binary", type=Path, required=True)
    parser.add_argument("--screenshots", type=Path)
    args = parser.parse_args()
    binary = args.binary.resolve()
    fixture = Path(__file__).parent / "fixtures" / "establishments.csv"
    if args.screenshots:
        args.screenshots.mkdir(parents=True, exist_ok=True)

    # A preview executable must never accidentally open a plaintext LAN listener.
    refused = subprocess.run([str(binary), "--listen", "192.0.2.10:8787"], capture_output=True, text=True, timeout=10)
    assert refused.returncode != 0 and "requiere tls-cert" in refused.stderr

    with tempfile.TemporaryDirectory(prefix="modding-browser-") as temporary, sync_playwright() as playwright:
        directory = Path(temporary)
        process, url, token = launch(binary, directory, fixture)
        browser = playwright.chromium.launch()
        context = browser.new_context(viewport={"width": 1440, "height": 1000})
        page = context.new_page()
        errors = []
        page.on("pageerror", lambda error: errors.append(str(error)))
        page.on("response", lambda response: errors.append(f"asset {response.status}") if "/assets/" in response.url and response.status >= 400 else None)
        try:
            login(page, url, token)
            expect(page.get_by_text("No hay un establecimiento seleccionado por defecto.", exact=False)).to_be_visible()
            page.get_by_label("Nombre, código DEIS o comuna").fill("arbol")
            page.get_by_role("button", name="Buscar", exact=True).click()
            page.get_by_role("button", name="Seleccionar Centro Árbol de Prueba").click()
            page.get_by_label("Nombre de dominio", exact=True).fill("gestion.example.cl")
            page.get_by_label("Sectores", exact=True).fill("Norte\nSur")
            page.get_by_label("Programas", exact=True).fill("Salud mental\nPromoción")
            page.get_by_label("Unidades", exact=True).fill("Dirección\nSOME")
            page.get_by_role("checkbox", name=re.compile("Farmacia")).check()
            page.get_by_role("button", name="Guardar y revisar").click()
            expect(page.get_by_role("heading", name="Revise su plan")).to_be_visible()
            expect(page.get_by_role("cell", name="Sectores/Norte", exact=True)).to_be_visible()
            page.get_by_role("button", name="Confirmar plan").click()
            expect(page.get_by_text("Plan confirmado y guardado.", exact=True)).to_be_visible()
            page.evaluate("document.fonts.ready")
            if args.screenshots:
                page.screenshot(path=str(args.screenshots / "wizard-desktop.png"), full_page=True)
            with page.expect_download() as download:
                page.get_by_role("link", name="Descargar configuración").click()
            payload = json.loads(Path(download.value.path()).read_text())
            assert payload["establishment"]["code"] == "900001"
            assert payload["status"] == "confirmed" and payload["optional_apps"] == ["farmacia"]
            assert payload["catalog_version"] == "synthetic-smoke-v1" and len(payload["catalog_sha256"]) == 64
            page.set_viewport_size({"width": 390, "height": 844})
            page.get_by_role("link", name="Editar configuración", exact=True).click()
            assert page.evaluate("document.documentElement.scrollWidth <= window.innerWidth + 1")
            if args.screenshots:
                page.screenshot(path=str(args.screenshots / "wizard-mobile.png"), full_page=True)
            page.get_by_role("button", name="Guardar y revisar").click()
            expect(page.get_by_role("button", name="Confirmar plan")).to_be_visible()
            assert page.evaluate("document.documentElement.scrollWidth <= window.innerWidth + 1")
            page.get_by_role("button", name="Confirmar plan").click()
            assert not errors, errors
            context.close()
            process.terminate()
            process.wait(timeout=5)
            process, url, token = launch(binary, directory, fixture, secure=True)
            # The generated certificate is only a disposable TLS transport fixture.
            context = browser.new_context(ignore_https_errors=True)
            page = context.new_page()
            login(page, url, token)
            cookies = context.cookies()
            assert len(cookies) == 1 and cookies[0]["secure"] and cookies[0]["httpOnly"]
            assert cookies[0]["sameSite"] == "Strict"
            page.goto(url + "/preview")
            expect(page.get_by_text("Plan confirmado y guardado.", exact=True)).to_be_visible()
            page.goto(url + "/establishments?q=900002")
            page.get_by_role("button", name="Seleccionar Posta Norte de Prueba").click()
            expect(page.get_by_label("Nombre de dominio", exact=True)).to_have_value("")
            expect(page.get_by_label("Sectores", exact=True)).to_have_value("")
            expect(page.get_by_role("checkbox", name=re.compile("Farmacia"))).not_to_be_checked()
            print("PASS: DEIS selection, saved preview, confirmation/export, TLS restart, secure cookies, isolation, mobile layout and LAN TLS guard.")
        finally:
            browser.close()
            if process.poll() is None:
                process.terminate()
                process.wait(timeout=5)


if __name__ == "__main__":
    main()
