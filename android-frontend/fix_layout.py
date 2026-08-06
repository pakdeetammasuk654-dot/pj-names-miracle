with open("app/src/main/java/com/miracle/names/ui/wizard/SearchWizardScreen.kt", "r") as f:
    content = f.read()

# 1. Remove .padding(16.dp) from the outer column
content = content.replace(".verticalScroll(scrollState)\n                .padding(16.dp),", ".verticalScroll(scrollState)\n                .padding(vertical = 16.dp),")

# 2. Add padding to the Text above LazyRow
content = content.replace("""                Text(
                    text = "💡 ลองเลือกชื่อคนดัง (Celebrity Avatars):",
                    fontSize = 12.sp,
                    color = TextSecondary
                )""", """                Text(
                    text = "💡 ลองเลือกชื่อคนดัง (Celebrity Avatars):",
                    fontSize = 12.sp,
                    color = TextSecondary,
                    modifier = Modifier.padding(horizontal = 16.dp)
                )""")

# 3. Remove layout hack from LazyRow and use simple fillMaxWidth()
old_lazy_row = """                androidx.compose.foundation.lazy.LazyRow(
                    state = listState,
                    modifier = Modifier
                        .fillMaxWidth()
                        .layout { measurable, constraints ->
                            val placeable = measurable.measure(
                                constraints.copy(
                                    maxWidth = constraints.maxWidth + 32.dp.roundToPx()
                                )
                            )
                            layout(placeable.width, placeable.height) {
                                placeable.place(-16.dp.roundToPx(), 0)
                            }
                        },
                    contentPadding = PaddingValues(horizontal = 16.dp),
                    horizontalArrangement = Arrangement.spacedBy(4.dp)
                ) {"""
new_lazy_row = """                androidx.compose.foundation.lazy.LazyRow(
                    state = listState,
                    modifier = Modifier.fillMaxWidth(),
                    contentPadding = PaddingValues(horizontal = 16.dp),
                    horizontalArrangement = Arrangement.spacedBy(4.dp)
                ) {"""
content = content.replace(old_lazy_row, new_lazy_row)

# 4. Wrap everything after LazyRow block in a padded Column
split_marker = "                Spacer(modifier = Modifier.height(8.dp))\n            }\n\n            Text("
if split_marker in content:
    parts = content.split(split_marker)
    
    wrapped_second_part = """                Spacer(modifier = Modifier.height(8.dp))
            }

            Column(
                modifier = Modifier.fillMaxWidth().padding(horizontal = 16.dp),
                verticalArrangement = Arrangement.spacedBy(16.dp)
            ) {
            Text(""" + parts[1]
            
    # Add a closing brace for the new Column just before the final Scaffold closing braces
    wrapped_second_part = wrapped_second_part.rsplit("        }\n    }\n}", 1)
    if len(wrapped_second_part) == 2:
        content = parts[0] + wrapped_second_part[0] + "            }\n        }\n    }\n}"
    else:
        print("Error: Could not find end of file structure")
else:
    print("Error: Could not find split marker")

with open("app/src/main/java/com/miracle/names/ui/wizard/SearchWizardScreen.kt", "w") as f:
    f.write(content)
print("Done")
