with open("app/src/main/java/com/miracle/names/ui/wizard/SearchWizardScreen.kt", "r") as f:
    lines = f.readlines()

depth = 0
for i, line in enumerate(lines[:550]):
    open_b = line.count('{')
    close_b = line.count('}')
    old_depth = depth
    depth += open_b
    depth -= close_b
    if depth != old_depth:
        print(f"Line {i+1}: {old_depth} -> {depth} | {line.strip()}")
