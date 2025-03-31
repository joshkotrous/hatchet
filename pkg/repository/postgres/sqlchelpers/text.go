package sqlchelpers

import "github.com/jackc/pgx/v5/pgtype"

func TextFromStr(str string) (pgtype.Text, error) {
	var pgText pgtype.Text

	if err := pgText.Scan(str); err != nil {
		return pgtype.Text{}, err
	}

	return pgText, nil
}