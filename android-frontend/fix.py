import re
with open("app/src/main/java/com/miracle/names/ui/blessing/BlessingCardScreen.kt", "r", encoding="utf-8") as f:
    content = f.read()

content = re.sub(
    r'val taksaInfo = .*?fontSize = 13\.sp\n\s*\)',
    r'''val taksaInfo = if (item.taksaChars.isNotEmpty() && item.taksaChars != "-") {
                            "📌 พลังอักขระทักษา: อักษร${item.taksaLabel} ${item.taksaCount} ตัว (${item.taksaChars})"
                        } else {
                            "📌 พลังอักขระทักษา: ไม่พบอักขระเป้าหมาย"
                        }
                        Text(
                            text = taksaInfo,
                            color = Color.White,
                            fontSize = 13.sp
                        )''',
    content,
    flags=re.DOTALL
)

content = re.sub(
    r'// Miracle Number Pair Analysis.*?\(Miracle Pairs\)",',
    r'''// Miracle Number Pair Analysis
                    Text(
                        text = "🔮 วิเคราะห์พลังคู่เลขมงคล (Miracle Pairs)",''',
    content,
    flags=re.DOTALL
)

with open("app/src/main/java/com/miracle/names/ui/blessing/BlessingCardScreen.kt", "w", encoding="utf-8") as f:
    f.write(content)
