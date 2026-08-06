import re

with open("app/src/main/java/com/miracle/names/ui/wizard/SearchWizardScreen.kt", "r", encoding="utf-8") as f:
    content = f.read()

# 1. Remove Gender Selector
gender_selector = """            // Gender Selector
            Text(
                text = "2. เลือกเพศ (Gender)",
                style = MaterialTheme.typography.titleMedium,
                fontWeight = FontWeight.Bold
            )
            val genderList = listOf(
                "all" to "👫 ทุกเพศ (ทั้งหมด)",
                "male" to "👨 เพศชาย",
                "female" to "👩 เพศหญิง"
            )
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.spacedBy(8.dp)
            ) {
                genderList.forEach { (key, label) ->
                    SelectableChip(
                        selected = gender == key,
                        text = label,
                        onClick = { viewModel.gender.value = key },
                        modifier = Modifier.weight(1f)
                    )
                }
            }

"""
content = content.replace(gender_selector, "")

# 2. Renumber Step 3 to 2
content = content.replace('text = "3. เลือกวันเกิด (🛡️ ปลอดกาลกิณี 100%)"', 'text = "2. เลือกวันเกิด (🛡️ ปลอดกาลกิณี 100%)"')

# 3. Renumber Step 4 to 3
content = content.replace('text = "4. คำอธิษฐานมงคล (AI Semantic Search)"', 'text = "3. คำอธิษฐานมงคล (AI Semantic Search)"')

with open("app/src/main/java/com/miracle/names/ui/wizard/SearchWizardScreen.kt", "w", encoding="utf-8") as f:
    f.write(content)
