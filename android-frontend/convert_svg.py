import re

with open("/Users/tayap/pj-names-miracle/go-backend/thai_buddha/sunday_buddha.svg", "r") as f:
    svg_data = f.read()

# Parse viewbox
vb_match = re.search(r'viewBox="([\d\.]+) ([\d\.]+) ([\d\.]+) ([\d\.]+)"', svg_data)
if not vb_match:
    print("No viewBox")
    exit(1)
w, h = float(vb_match.group(3)), float(vb_match.group(4))

# Parse styles
style_matches = re.findall(r'\.cls-(\d+) {\s*fill: (#\w+);\s*}', svg_data)
color_map = {f"cls-{k}": v for k, v in style_matches}

# Parse paths
paths = re.findall(r'<path class="(cls-\d+)" d="([^"]+)"', svg_data)

out_xml = f"""<?xml version="1.0" encoding="utf-8"?>
<vector xmlns:android="http://schemas.android.com/apk/res/android"
    android:width="24dp"
    android:height="{24*(h/w):.2f}dp"
    android:viewportWidth="{w}"
    android:viewportHeight="{h}">
"""

for cls, d in paths:
    fillColor = color_map.get(cls, "#FFFFFF")
    out_xml += f'    <path android:fillColor="{fillColor}" android:pathData="{d}"/>\n'

out_xml += "</vector>"

with open("/Users/tayap/pj-names-miracle/android-frontend/app/src/main/res/drawable/ic_buddha_sunday.xml", "w") as f:
    f.write(out_xml)

print("Conversion complete.")
