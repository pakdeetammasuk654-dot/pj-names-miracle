import re

with open("repository/name_repository.go", "r", encoding="utf-8") as f:
    content = f.read()

# 1. Update Interface
content = content.replace(
    "SearchSemanticWishNames(ctx context.Context, wishQuery string, dayOfBirth string, gender string, limit int) ([]domain.NameResultItem, error)",
    "SearchSemanticWishNames(ctx context.Context, targetSatSums []int, targetShaSums []int, wishQuery string, dayOfBirth string, gender string, limit int) ([]domain.NameResultItem, error)"
)

# 2. Update Implementation Signature
content = content.replace(
    "func (r *nameRepository) SearchSemanticWishNames(ctx context.Context, wishQuery string, dayOfBirth string, gender string, limit int) ([]domain.NameResultItem, error) {",
    "func (r *nameRepository) SearchSemanticWishNames(ctx context.Context, targetSatSums []int, targetShaSums []int, wishQuery string, dayOfBirth string, gender string, limit int) ([]domain.NameResultItem, error) {"
)

# 3. Update SQL Query inside SearchSemanticWishNames
old_sql = """		vectorQuery := fmt.Sprintf(`
			SELECT 
				name_id, 
				thname, 
				COALESCE(meaning, ''), 
				sat_sum, 
				sha_sum,
				(meaning_vector <=> $1::vector) as distance
			FROM names_miracle 
			WHERE meaning_vector IS NOT NULL %s
			ORDER BY meaning_vector <=> $1::vector ASC
			LIMIT $2
		`, dayClause)

		rows, err := r.db.QueryContext(ctx, vectorQuery, vecStr, 1000)"""

new_sql = """		vectorQuery := fmt.Sprintf(`
			SELECT 
				name_id, 
				thname, 
				COALESCE(meaning, ''), 
				sat_sum, 
				sha_sum,
				(meaning_vector <=> $1::vector) as distance
			FROM names_miracle 
			WHERE meaning_vector IS NOT NULL 
			AND sat_sum = ANY($3) AND sha_sum = ANY($4)
			%s
			ORDER BY meaning_vector <=> $1::vector ASC
			LIMIT $2
		`, dayClause)

		rows, err := r.db.QueryContext(ctx, vectorQuery, vecStr, 1000, pq.Array(targetSatSums), pq.Array(targetShaSums))"""

content = content.replace(old_sql, new_sql)

with open("repository/name_repository.go", "w", encoding="utf-8") as f:
    f.write(content)
