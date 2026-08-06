import re

with open("usecase/pipeline/mode_filter.go", "r", encoding="utf-8") as f:
    content = f.read()

old_logic = """	if payload.Request.WishQuery != "" {
		cands, err := f.repo.SearchSemanticWishNames(ctx, payload.Request.WishQuery, payload.Request.DayOfBirth, payload.Request.Gender, fetchLimit)
		if err == nil && len(cands) > 0 {
			candidates = cands
		} else if err != nil {
			log.Printf("[Pipeline Step 1: ModeFilter] SearchSemanticWishNames error: %v", err)
		}
	}

	if len(candidates) == 0 {
		goodSumList := make([]int, 0, len(goodMap))
		for num := range goodMap {
			goodSumList = append(goodSumList, num)
		}
		sort.Ints(goodSumList)

		cands, err := f.repo.SearchCandidateNames(ctx, goodSumList, goodSumList, payload.Request.DayOfBirth, payload.Request.Gender, payload.Request.WishQuery, fetchLimit)"""

new_logic = """	goodSumList := make([]int, 0, len(goodMap))
	for num := range goodMap {
		goodSumList = append(goodSumList, num)
	}
	sort.Ints(goodSumList)

	if payload.Request.WishQuery != "" {
		cands, err := f.repo.SearchSemanticWishNames(ctx, goodSumList, goodSumList, payload.Request.WishQuery, payload.Request.DayOfBirth, payload.Request.Gender, fetchLimit)
		if err == nil && len(cands) > 0 {
			candidates = cands
		} else if err != nil {
			log.Printf("[Pipeline Step 1: ModeFilter] SearchSemanticWishNames error: %v", err)
		}
	}

	if len(candidates) == 0 {
		cands, err := f.repo.SearchCandidateNames(ctx, goodSumList, goodSumList, payload.Request.DayOfBirth, payload.Request.Gender, payload.Request.WishQuery, fetchLimit)"""

content = content.replace(old_logic, new_logic)

with open("usecase/pipeline/mode_filter.go", "w", encoding="utf-8") as f:
    f.write(content)
