with open("app/src/main/java/com/miracle/names/ui/wizard/SearchWizardScreen.kt", "r") as f:
    lines = f.readlines()

depth = 0
for i, line in enumerate(lines):
    open_b = line.count('{')
    close_b = line.count('}')
    depth += open_b
    depth -= close_b
    if depth == 0 and open_b > 0 or close_b > 0:
        print(f"Line {i+1}: Depth reached 0! Line content: {line.strip()}")
        
print(f"Final depth: {depth}")
