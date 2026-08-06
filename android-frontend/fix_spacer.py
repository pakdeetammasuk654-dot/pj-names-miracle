import re

with open("app/src/main/java/com/miracle/names/ui/results/NameResultsScreen.kt", "r", encoding="utf-8") as f:
    content = f.read()

double_spacer = """                    Spacer(modifier = Modifier.weight(1f))
                    Spacer(modifier = Modifier.weight(1f))"""

single_spacer = """                    Spacer(modifier = Modifier.weight(1f))"""

content = content.replace(double_spacer, single_spacer)

with open("app/src/main/java/com/miracle/names/ui/results/NameResultsScreen.kt", "w", encoding="utf-8") as f:
    f.write(content)
