# Third-Party Notices and Attribution

This project incorporates open source software. We thank the authors and communities that make this work possible.

This `ais-demo` module is a small Go web app that integrates with Ollama. It does not bundle third‑party JS build systems; it references a few browser CDN assets and uses Go libraries through the Go module system.

## Direct Runtime Dependencies

- Ollama (server) — pulled as a Docker image and accessed via HTTP API.
  - Image: `ollama/ollama`
  - Site: https://ollama.com/
  - License: see upstream image repository

- CodeMirror (CDN, browser runtime only) — inline code editor used for UIA/APA/APr and audit views.
  - Site: https://codemirror.net/
  - License: MIT
  - Referenced via jsDelivr CDN in the HTML.

## Go Modules (indirect and direct)

This module uses only the Go standard library and the following first‑party local packages:

- `ais-demo/internal/ais` (local): signing (demo JWS HS256), guard, ollama client, http tool, types.

If additional Go modules are added, run:

```bash
go list -m -json all > third_party_go_modules.json
```

and update this file with their licenses/attributions.

## Docker Compose and Local Models

- This repository mounts `ais-demo/internal/ollama_models` into the Ollama container at `/root/.ollama`. Ensure any model weights and licenses you include are compliant with their upstream terms.

## Notices

- Some artifacts are pulled at runtime from third‑party sources (Ollama images, CodeMirror via CDN). Their licenses apply. This project does not redistribute those artifacts.
- If you notice any attribution missing or inaccurate, please open an issue or PR.
