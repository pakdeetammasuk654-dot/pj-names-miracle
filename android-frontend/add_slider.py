import re

with open("app/src/main/java/com/miracle/names/ui/wizard/SearchWizardScreen.kt", "r", encoding="utf-8") as f:
    content = f.read()

# Add state collection
content = content.replace(
    "val filterTwoSyllables by viewModel.filterTwoSyllables.collectAsState()",
    "val filterTwoSyllables by viewModel.filterTwoSyllables.collectAsState()\n    val semanticWeight by viewModel.semanticWeight.collectAsState()"
)

slider_ui = """
            // Slider for Weight
            Spacer(modifier = Modifier.height(8.dp))
            Text(
                text = "น้ำหนักการจัดอันดับ (Ranking Weight)",
                style = MaterialTheme.typography.bodyMedium,
                fontWeight = FontWeight.Bold
            )
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween
            ) {
                Text(text = "เน้นเลขศาสตร์", fontSize = 12.sp, color = TextSecondary)
                Text(text = "สมดุล", fontSize = 12.sp, color = TextSecondary)
                Text(text = "เน้นความหมาย", fontSize = 12.sp, color = TextSecondary)
            }
            Slider(
                value = semanticWeight,
                onValueChange = { viewModel.semanticWeight.value = it },
                valueRange = 0f..1f,
                steps = 9,
                modifier = Modifier.fillMaxWidth()
            )
            val weightPercent = (semanticWeight * 100).toInt()
            Text(
                text = "ความหมาย: ${weightPercent}%, เลขศาสตร์: ${100 - weightPercent}%",
                fontSize = 12.sp,
                color = GoldPrimary,
                modifier = Modifier.align(Alignment.CenterHorizontally)
            )

            OutlinedTextField(
"""

content = content.replace("            OutlinedTextField(", slider_ui.strip() + "\n\n            OutlinedTextField(")

# Ensure Slider is imported
if "import androidx.compose.material3.Slider" not in content:
    content = content.replace("import androidx.compose.material3.*", "import androidx.compose.material3.*\nimport androidx.compose.material3.Slider")

with open("app/src/main/java/com/miracle/names/ui/wizard/SearchWizardScreen.kt", "w", encoding="utf-8") as f:
    f.write(content)
