with open("app/src/main/java/com/miracle/names/ui/wizard/SearchWizardScreen.kt", "r") as f:
    lines = f.readlines()

depth = 0
for i, line in enumerate(lines):
    open_b = line.count('{')
    close_b = line.count('}')
    old_depth = depth
    depth += open_b
    depth -= close_b
    if old_depth > 0 and depth == 0:
        print(f"Line {i+1}: Block closed. Line content: {line.strip()}")
