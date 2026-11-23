# Uchinchi tomon komponentlari (SBOM)

Mashina o‘qiydigan nusxa: [`sbom.cdx.json`](sbom.cdx.json) (CycloneDX 1.5).

| Komponent | Versiya | Litsenziya | Vazifasi |
|---|---|---|---|
| go-webview2 | 0.0.0-20260205173254-56598839c808 | MIT | Dastur oynasi (Microsoft WebView2 bilan bog‘lanish) |
| go-winloader | 0.0.0-20250406163304-c1995be93bd1 | MIT | WebView2Loader yuklovchi |
| golang.org/x/sys (github.com/golang/sys) | 0.30.0 | BSD-3-Clause | Windows tizim chaqiruvlari, registr |
| Go standart kutubxonasi | 1.24.7 | BSD-3-Clause | HTTP, kriptografiya (Ed25519, SHA-256) |
| JSZip | 3.10.1 | MIT OR GPL-3.0 | docx/xlsx/pptx arxivlarini ochish |
| hunspell-asm | 4.0.2 | MIT | Hunspell (WebAssembly) imlo dvigateli |
| Hunspell | fc6a952-200131 | MPL-1.1 OR GPL-2.0-or-later OR LGPL-2.1-or-later | hunspell-asm ichidagi imlo dvigateli |
| emscripten-wasm-loader | 3.0.3 | MIT | hunspell-asm ichida WASM yuklovchi |
| IBM Plex Sans / Mono (Fontsource) | 5.1.1 | OFL-1.1 | Shrift |
| bitg-hunspell uz_UZ-2026 | 0.1.1 | GPL-3.0-only | Imlo lug‘ati (2026 alifbosi) |
| uzb-frequent-words | master | GPL-3.0-only | So‘zlar chastotasi (kirill → 2026 lotin) |

**Izohlar**
- `nanoid@2.1.11` (hunspell-asm bog‘liqligi) zaiflik e’lonlari (GHSA-mwcw-c2x4-8c55, GHSA-28wg-ghj8-5hjv va boshq.) sababli
  yig‘ishda `src/web/nanoid-shim.js` bilan almashtirilgan — dastur ichida nanoid kodi yo‘q.
- `esbuild` faqat yig‘ish vositasi (0.25.0, GHSA-67mh-4wv8-2f99 tuzatilgan), dastur tarkibiga kirmaydi.
- Microsoft Edge WebView2 Runtime — Windows tarkibiy qismi, Microsoft tomonidan yangilanadi, dastur bilan tarqatilmaydi.
- hunspell-asm 2020-yildan beri yangilanmagan. U faqat foydalanuvchining o‘z matnini va imzolangan lug‘atni WebView ichidagi
  WebAssembly qum qutisida (sandbox) qayta ishlaydi; tarmoqqa va fayllarga to‘g‘ridan-to‘g‘ri kira olmaydi.
