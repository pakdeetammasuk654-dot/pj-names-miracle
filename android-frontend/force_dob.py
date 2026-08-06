import re

with open("app/src/main/java/com/miracle/names/ui/wizard/SearchWizardScreen.kt", "r", encoding="utf-8") as f:
    content = f.read()

# 1. Remove the "ทุกวัน" option
content = content.replace('"" to "ทุกวัน (ไม่ระบุ)",\n', '')
content = content.replace('"" to "ทุกวัน (ไม่ระบุ)",\r\n', '')

# 2. Disable search button if dayOfBirth is empty
old_btn = """            Button(
                onClick = { viewModel.searchNames() },
                modifier = Modifier
                    .fillMaxWidth()
                    .height(54.dp),
                shape = RoundedCornerShape(12.dp),
                colors = ButtonDefaults.buttonColors(containerColor = GoldPrimary)
            ) {"""

new_btn = """            Button(
                onClick = { viewModel.searchNames() },
                enabled = dayOfBirth.isNotEmpty(),
                modifier = Modifier
                    .fillMaxWidth()
                    .height(54.dp),
                shape = RoundedCornerShape(12.dp),
                colors = ButtonDefaults.buttonColors(
                    containerColor = GoldPrimary,
                    disabledContainerColor = Color.LightGray
                )
            ) {"""

content = content.replace(old_btn, new_btn)

with open("app/src/main/java/com/miracle/names/ui/wizard/SearchWizardScreen.kt", "w", encoding="utf-8") as f:
    f.write(content)
