import re

with open("app/src/main/java/com/miracle/names/ui/wizard/SearchWizardScreen.kt", "r", encoding="utf-8") as f:
    content = f.read()

# Fix the first broken OutlinedTextField (Surname)
broken_surname = """    // Slider for Weight
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

            OutlinedTextField("""

fixed_surname = """            OutlinedTextField("""

content = content.replace(broken_surname, fixed_surname)

# Fix the second broken OutlinedTextField (wishQuery)
broken_wish = """// Slider for Weight
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

            OutlinedTextField("""

fixed_wish = """            // Slider for Weight
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

            Spacer(modifier = Modifier.height(8.dp))
            OutlinedTextField("""

content = content.replace(broken_wish, fixed_wish)

with open("app/src/main/java/com/miracle/names/ui/wizard/SearchWizardScreen.kt", "w", encoding="utf-8") as f:
    f.write(content)
