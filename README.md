# Yangi alifbo konvertori

Word, Excel va PowerPoint hujjatlaridagi o‘zbekcha matnni O‘RQ–1184-son qonun bilan
tasdiqlangan **2026-yilgi lotin alifbosiga** (Ö, Ğ, Ş, Ç) o‘giradigan va imlo xatolarini
topadigan Windows dasturi.

## Yuklab olish

[**Eng so‘nggi versiya → Yangi_alifbo_konvertori_setup.exe**](https://github.com/otashnst/yangi-alifbo-konvertori/raw/main/release/Yangi_alifbo_konvertori_setup.exe)

Talab: Windows 10 yoki 11. O‘rnatish uchun administrator paroli kerak emas.
Birinchi ishga tushirishda Windows «Windows protected your PC» deb ogohlantirsa:
**More info → Run anyway**.

## Imkoniyatlar

- `.docx`, `.xlsx`, `.pptx`, `.txt` fayllar: shrift, jadval va rasmlar o‘zgarmaydi
- Eski lotin (O‘, G‘, Sh, Ch) va kirill yozuvidan yangi alifboga
- Imlo tekshiruvi: apostrofsiz yozilgan so‘zlar, x/h xatolari va boshqalar — har bir tuzatishni foydalanuvchi tasdiqlaydi
- Internetsiz ishlaydi, hujjatlar kompyuterdan tashqariga chiqmaydi
- Imlo lug‘ati ([yangi-alifbo-lugat](https://github.com/otashnst/yangi-alifbo-lugat)) va dasturning o‘zi avtomatik yangilanadi

## Yangilanishlar xavfsizligi

`release/` papkasida o‘rnatuvchi va `latest.json` turadi: unda o‘rnatuvchining SHA-256 summasi va
muallifning Ed25519 imzosi bor. Dastur imzoni ichiga joylangan ochiq kalit bilan tekshiradi
va imzosiz yoki o‘zgartirilgan yangilanishni o‘rnatmaydi.

## Yangi versiya chiqarish

1. Yangi `Yangi_alifbo_konvertori_setup.exe` ni maxfiy kalit bilan imzolang:
   `sign_release.exe release_private_key.txt Yangi_alifbo_konvertori_setup.exe 1.8.0 "Nima o‘zgardi" latest.json`
2. Ikkala faylni `release/` papkasidagi eskilarining o‘rniga yuklang (commit).

Bir necha daqiqadan keyin barcha kompyuterlardagi dastur yangi versiyani taklif qiladi.

## Muallif

Abduxoliqov Otabek · 2026 · (33) 071-60-70

Imlo lug‘ati: bitg-hunspell (uz-hunspell asosida) va uzb-frequent-words — GPL-3.0.
Shrift: IBM Plex — SIL Open Font License.
