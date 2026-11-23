#!/usr/bin/env bash
# Reproducible build on Linux: Go 1.24, Node 22, NSIS 3.09, mingw-w64 binutils (windres), Python 3.
set -euo pipefail
cd "$(dirname "$0")"
# 1. dictionary (from the yangi-alifbo-lugat repository)
mkdir -p app/dict
for f in uz_UZ-2026.aff uz_UZ-2026.dic freq.txt; do
  curl -fsSL "https://raw.githubusercontent.com/otashnst/yangi-alifbo-lugat/main/$f" -o "app/dict/$f"
done
curl -fsSL https://raw.githubusercontent.com/otashnst/yangi-alifbo-lugat/main/LICENSE -o app/dict/LICENSE-uz_UZ-2026-GPL3.txt
# 2. web bundle
(cd web && npm ci --ignore-scripts 2>/dev/null || npm install --ignore-scripts)
(cd web && npx esbuild entry.js --bundle --platform=browser --format=iife --minify --alias:nanoid=./nanoid-shim.js --outfile=hunspell.bundle.js)
python3 web/build_page.py
# 3. Windows exe
(cd app && x86_64-w64-mingw32-windres --preprocessor=cpp -O coff -F pe-x86-64 res.rc -o rsrc_windows_amd64.syso)
(cd app && GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -ldflags "-s -w -H windowsgui" -o ../installer/Yangi_alifbo_konvertori.exe .)
# 4. installer
cp app/dict/LICENSE-uz_UZ-2026-GPL3.txt installer/ && cp app/fonts/LICENSE-IBM-Plex-OFL.txt installer/
# uzb-frequent-words is GPL-3.0 as well (same license text)
cp app/dict/LICENSE-uz_UZ-2026-GPL3.txt installer/LICENSE-freq-GPL3.txt
(cd installer && makensis -V2 setup.nsi)
sha256sum installer/Yangi_alifbo_konvertori.exe installer/Yangi_alifbo_konvertori_setup.exe
