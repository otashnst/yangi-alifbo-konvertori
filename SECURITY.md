# Xavfsizlik — Yangi alifbo konvertori

Bu hujjat axborot xavfsizligi bo‘limlari va auditorlar uchun: dastur nima qiladi, qayerga murojaat qiladi,
qanday himoyalangan va qanday cheklovlari bor.

## 1. Ma’lumotlar bilan ishlash

- Hujjatlar faqat foydalanuvchi kompyuterida, dastur oynasi ichida qayta ishlanadi.
  **Hujjat matni, fayl nomlari yoki mazmuni hech qayerga yuborilmaydi.** Telemetriya, statistika, reklama yo‘q.
- Natija faqat foydalanuvchi tanlagan joyga (Windows «Saqlash» oynasi) yoki
  `Documents\Yangi alifbo` papkasiga yoziladi. Fayl nomidan papka qismlari va `..` olib tashlanadi,
  faqat `.docx .xlsx .pptx .txt` kengaytmalari ruxsat etilgan.
- Dastur yozadigan joylar:

| Joy | Nima |
|---|---|
| `%LOCALAPPDATA%\Programs\Yangi alifbo konvertori` | Dasturning o‘zi |
| `%LOCALAPPDATA%\YangiAlifbo\log.txt` | Texnik jurnal (fayl nomlari va matn **yozilmaydi**), har ishga tushishda tozalanadi |
| `%LOCALAPPDATA%\YangiAlifbo\settings.json` | Foydalanuvchi sozlamasi (internetsiz rejim) |
| `%LOCALAPPDATA%\YangiAlifbo\dict\` | Yangilangan imlo lug‘ati (imzosi tekshirilgan) |
| `%LOCALAPPDATA%\YangiAlifbo\WebView2\` | WebView2 ish papkasi |
| `%LOCALAPPDATA%\YangiAlifbo\ai\` | AI tahlilchi: Gemini kaliti (Windows DPAPI bilan shifrlangan) va tanlangan model |

## 2. Tarmoq

Dastur o‘zi faqat ikkita manzilga, faqat **o‘qish** uchun murojaat qiladi:

| Manzil | Nima uchun |
|---|---|
| `https://raw.githubusercontent.com/otashnst/yangi-alifbo-lugat/main/` | Imlo lug‘ati yangilanishi |
| `https://raw.githubusercontent.com/otashnst/yangi-alifbo-konvertori/main/release/` | Dastur yangilanishi |

Uchinchi manzil — `https://generativelanguage.googleapis.com/` (Google Gemini) — faqat **AI tahlilchi** uchun,
faqat foydalanuvchi o‘z kalitini kiritgan va «Tahlil qilish» tugmasini bosgan paytda ishlatiladi (5-band).

**Internetni butunlay o‘chirish** (yopiq tarmoqlar uchun):
- foydalanuvchi: chap paneldagi «Internetga ulanmaslik» belgisi;
- administrator (foydalanuvchi o‘chira olmaydi, Group Policy orqali tarqatish mumkin):
  `HKLM\SOFTWARE\Policies\YangiAlifbo` → `Offline` (DWORD) = `1`
  (yoki `HKCU\...` — bitta foydalanuvchi uchun). Tayyor fayl: [`policy/offline.reg`](policy/offline.reg).

Bu rejimda dastur hech qanday tarmoq so‘rovi yubormaydi.

## 3. Ichki server

Dastur oynasi (Microsoft WebView2) interfeysni kompyuter ichidagi kichik serverdan oladi:
- faqat `127.0.0.1`, tasodifiy port — tarmoqdan ko‘rinmaydi;
- har ishga tushishda yangi 256-bitli token; tokensiz so‘rovlar `403` bilan rad etiladi
  (boshqa dasturlar va brauzerdagi saytlar unga murojaat qila olmaydi);
- `Host` sarlavhasi tekshiriladi (DNS rebinding himoyasi);
- Content-Security-Policy: tashqi skript, kadr va tarmoq manbalari taqiqlangan.

## 4. Yangilanishlar

- Dastur va lug‘at yangilanishlari muallifning **Ed25519** kaliti bilan imzolanadi; ochiq kalit dastur ichiga joylangan.
  Imzosi yo‘q yoki noto‘g‘ri yangilanish o‘rnatilmaydi — GitHub akkaunti buzilgan taqdirda ham.
- Har bir fayl SHA-256 va hajmi bo‘yicha tekshiriladi; yarim yuklangan lug‘at ishlatilmaydi.
- Maxfiy kalit muallif kompyuterida `relizchi` vositasi bilan yaratilgan va faqat u yerda, oflayn saqlanadi (`src/tools/relizchi`).
- Ochiq kalit (1.9+): `KMIIzV7Yg4m7Haig88pZoUQXjOQhv0Mkst7ENdjMVHk=` (1.8 dagi `ttgk…` kaliti ishlatilmaydi — 1.8 o‘rnatilgan kompyuterlarga 1.9 bir marta qo‘lda o‘rnatiladi)
- Dastur yangilanishi foydalanuvchi «Yangilash» tugmasini bosgandagina o‘rnatiladi.

## 5. AI tahlilchi (ixtiyoriy)

- O‘rnatilganda o‘chiq: foydalanuvchi o‘zining bepul Google Gemini API kalitini kiritmaguncha hech narsa yuborilmaydi.
- Kalit kiritishdan oldin foydalanuvchi ogohlantirishni tasdiqlaydi: **Gemini bepul tarifida Google yuborilgan matnlardan
  mahsulotlarini yaxshilash uchun foydalanishi va ularni xodimlari ko‘rib chiqishi mumkin.** Maxfiy hujjatlar uchun mo‘ljallanmagan.
- Matn faqat «Tahlil qilish» bosilganda, faqat Google’ga yuboriladi; fayllarni o‘girish va imlo tekshiruvi bundan mustaqil, internetsiz.
- Kalit Windows DPAPI bilan joriy foydalanuvchi uchun shifrlanadi, interfeys sahifasiga hech qachon berilmaydi, jurnalga yozilmaydi.
- Gemini javobi faqat matn sifatida ko‘rsatiladi (HTML/skript bajarilmaydi).
- **Administrator o‘chirishi:** `HKLM\SOFTWARE\Policies\YangiAlifbo` → `DisableAI` (DWORD) = `1` —
  tayyor fayl [`policy/disable-ai.reg`](policy/disable-ai.reg). `Offline` siyosati ham AI tahlilchini o‘chiradi.

## 6. Huquqlar

- Administrator huquqi talab qilinmaydi; xizmat (service), avtoyuklash, drayver o‘rnatilmaydi.
- O‘chirish: «Programs and Features» — barcha fayllar va `%LOCALAPPDATA%\YangiAlifbo` o‘chiriladi.

## 7. Manba kodi va komponentlar

- To‘liq manba kodi: [`src/`](src/), yig‘ish: [`src/build.sh`](src/build.sh).
- Uchinchi tomon komponentlari va litsenziyalar: [`THIRD_PARTY.md`](THIRD_PARTY.md), [`sbom.cdx.json`](sbom.cdx.json).

## 8. Ma’lum cheklovlar

- `.exe` Authenticode raqamli sertifikati bilan imzolanmagan — Windows SmartScreen ogohlantiradi.
- Imlo dvigateli `hunspell-asm` 2020-yildan beri yangilanmagan (qum qutisida ishlaydi, THIRD_PARTY.md).
- 1.7 va undan oldingi versiyalar eski reliz kalitiga ishonadi — barcha kompyuterlarni 1.8 ga yangilang.
- Dastur mustaqil tashkilot tomonidan sertifikatlanmagan.
- AI tahlilchi tashqi xizmatga (Google) bog‘liq; uning ma’lumotlar siyosati Google shartlari bilan belgilanadi.

## 9. Zaiflik haqida xabar berish

GitHub’da «Issues» orqali (maxfiy tafsilotlarsiz) yoki muallifga: Abduxoliqov Otabek, (33) 071-60-70.
