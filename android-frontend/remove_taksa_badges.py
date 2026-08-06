import re

# 1. Remove from NameResultsScreen.kt
with open("app/src/main/java/com/miracle/names/ui/results/NameResultsScreen.kt", "r", encoding="utf-8") as f:
    content1 = f.read()

content1 = re.sub(
    r'\s*// Taksa Badge\s*Box\(\s*modifier = Modifier\s*\.clip\(RoundedCornerShape\(6\.dp\)\)\s*\.background\(GoldLight\)\s*\.padding\(horizontal = 6\.dp, vertical = 3\.dp\)\s*\)\s*\{\s*val taksaDetail = if \(item\.taksaChars\.isNotEmpty\(\) && item\.taksaChars != "-"\) \{\s*"🌟\$\{item\.taksaLabel\}: \$\{item\.taksaCount\} ตัว \(\$\{item\.taksaChars\}\)"\s*\} else \{\s*"🌟\$\{item\.taksaLabel\}: \$\{item\.taksaCount\} ตัว"\s*\}\s*Text\(\s*text = taksaDetail,\s*fontSize = 11\.sp,\s*fontWeight = FontWeight\.Bold,\s*color = TextGold,\s*maxLines = 1,\s*softWrap = false\s*\)\s*\}\s*',
    '\n                    ',
    content1,
    flags=re.DOTALL
)

with open("app/src/main/java/com/miracle/names/ui/results/NameResultsScreen.kt", "w", encoding="utf-8") as f:
    f.write(content1)

# 2. Remove from BlessingCardScreen.kt
with open("app/src/main/java/com/miracle/names/ui/blessing/BlessingCardScreen.kt", "r", encoding="utf-8") as f:
    content2 = f.read()

content2 = re.sub(
    r'\s*val taksaInfo = if \(item\.taksaChars\.isNotEmpty\(\) && item\.taksaChars != "-"\) \{\s*"📌 พลังอักขระทักษา: อักษร\$\{item\.taksaLabel\} \$\{item\.taksaCount\} ตัว \(\$\{item\.taksaChars\}\)"\s*\} else \{\s*"📌 พลังอักขระทักษา: ไม่พบอักขระเป้าหมาย"\s*\}\s*Text\(\s*text = taksaInfo,\s*color = Color\.White,\s*fontSize = 13\.sp\s*\)\s*',
    '\n                    ',
    content2,
    flags=re.DOTALL
)

with open("app/src/main/java/com/miracle/names/ui/blessing/BlessingCardScreen.kt", "w", encoding="utf-8") as f:
    f.write(content2)

