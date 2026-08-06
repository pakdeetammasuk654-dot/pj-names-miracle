import re
with open("app/src/main/java/com/miracle/names/ui/blessing/BlessingCardScreen.kt", "r", encoding="utf-8") as f:
    content = f.read()

content = re.sub(
    r'fun formatDayThai\(day: String\): String \{.*?\n\}',
    r'''fun formatDayThai(day: String): String {
    return when (day.lowercase()) {
        "sunday" -> "วันอาทิตย์"
        "monday" -> "วันจันทร์"
        "tuesday" -> "วันอังคาร"
        "wednesday1" -> "วันพุธ (กลางวัน)"
        "wednesday2" -> "วันพุธ (กลางคืน)"
        "thursday" -> "วันพฤหัสบดี"
        "friday" -> "วันศุกร์"
        "saturday" -> "วันเสาร์"
        else -> "ทุกวัน (ไม่ระบุ)"
    }
}''',
    content,
    flags=re.DOTALL
)

with open("app/src/main/java/com/miracle/names/ui/blessing/BlessingCardScreen.kt", "w", encoding="utf-8") as f:
    f.write(content)
