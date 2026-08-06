import re

with open("usecase/pipeline/scoring_filter.go", "r", encoding="utf-8") as f:
    content = f.read()

# 1. Add FinalScore calculation inside the processing loop
calc_code = """
		// Calculate Hybrid FinalScore
		var semWeight float64 = 0.5
		if payload.Request.SemanticWeight != nil {
			semWeight = *payload.Request.SemanticWeight
		}
		numWeight := 1.0 - semWeight

		numerologyRaw := float64(item.TotalScore) / 2.0
		semanticRaw := 0.0
		if payload.Request.WishQuery != "" {
			semanticRaw = (1.0 - item.VectorDistance) * 100.0
			if semanticRaw < 0 {
				semanticRaw = 0
			}
		} else {
			// If no wish query, numerology dictates 100% of the score implicitly, but we just use numWeight
			semanticRaw = 0
			// Force numWeight to 1.0 if no semantic wish
			numWeight = 1.0
			semWeight = 0.0
		}
		
		item.FinalScore = (numerologyRaw * numWeight) + (semanticRaw * semWeight)

		processed = append(processed, item)
"""
content = content.replace("processed = append(processed, item)", calc_code.strip())

# 2. Update the sorting logic
old_sort = """
	// Smart Hybrid Ranking Strategy:
	// Priority #1: IsPerfectCombined (Descending) - Names where (Name + Surname) equals perfect D-type pair numbers float to top
	// Priority #2: TotalScore (Descending) - High numerology score
	// Priority #3: VectorDistance (Ascending) - AI Semantic Wish Distance (if wish query provided)
	hasWish := payload.Request.WishQuery != ""

	sort.Slice(processed, func(i, j int) bool {
		if processed[i].IsPerfectCombined != processed[j].IsPerfectCombined {
			return processed[i].IsPerfectCombined
		}
		if processed[i].TotalScore != processed[j].TotalScore {
			return processed[i].TotalScore > processed[j].TotalScore
		}
		if hasWish && processed[i].VectorDistance != processed[j].VectorDistance {
			return processed[i].VectorDistance < processed[j].VectorDistance
		}
		return false
	})
"""

new_sort = """
	// Smart Hybrid Ranking Strategy:
	// Priority #1: IsPerfectCombined (Descending) - Names where (Name + Surname) equals perfect D-type pair numbers float to top
	// Priority #2: FinalScore (Descending) - Hybrid weighted score (Numerology vs Semantic)

	sort.Slice(processed, func(i, j int) bool {
		if processed[i].IsPerfectCombined != processed[j].IsPerfectCombined {
			return processed[i].IsPerfectCombined
		}
		if processed[i].FinalScore != processed[j].FinalScore {
			return processed[i].FinalScore > processed[j].FinalScore
		}
		return processed[i].TotalScore > processed[j].TotalScore
	})
"""

content = content.replace(old_sort.strip(), new_sort.strip())

with open("usecase/pipeline/scoring_filter.go", "w", encoding="utf-8") as f:
    f.write(content)
