import re
with open("app/src/main/java/com/miracle/names/ui/blessing/BlessingCardScreen.kt", "r", encoding="utf-8") as f:
    content = f.read()

content = re.sub(
    r'text = "犧ｪ犹謂ｸ.*?吭ｸｲ犧｡",',
    r'text = "แชร์ / บันทึกการ์ดมงคลวิเคราะห์นาม",',
    content,
    flags=re.DOTALL
)

with open("app/src/main/java/com/miracle/names/ui/blessing/BlessingCardScreen.kt", "w", encoding="utf-8") as f:
    f.write(content)
