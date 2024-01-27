package scanalert_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	scanalert "github.com/tscolari/gormalert"
	"github.com/tscolari/gormalert/testenv"
)

func Test_ScanAlert(t *testing.T) {
	testCases := map[string]struct {
		table       string
		where       string
		shouldAlert bool
	}{
		"using primary key": {
			table:       "fruits",
			where:       "id = 1",
			shouldAlert: false,
		},

		"using name index": {
			table:       "vegetables",
			where:       "name = 'potato'",
			shouldAlert: false,
		},

		"searching for name without index": {
			table:       "fruits",
			where:       "name != 'apple'",
			shouldAlert: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			r := require.New(t)
			db := testenv.InitPgDB(t)

			var alerted bool
			var explainResult string

			scanalert.RegisterScanAlert(db, scanalert.DefaultAlertOptions(), func(source string, result string) {
				explainResult = result
				alerted = true
			})

			var result struct{}

			err := db.Table(tc.table).Where(tc.where).Find(&result).Error
			r.NoError(err)

			if tc.shouldAlert {
				r.True(alerted, "should have alerted:\n%s", explainResult)
			} else {
				r.False(alerted, "should not have alerted\n%s", explainResult)
			}

		})
	}
}

func Test_ScanAlert_WithRaw(t *testing.T) {
	testCases := map[string]struct {
		query       string
		shouldAlert bool
	}{
		"SELECT using primary key": {
			query:       "SELECT * FROM fruits WHERE id = 1",
			shouldAlert: false,
		},

		"SELECT using non-indexed name": {
			query:       "SELECT * FROM fruits WHERE name != 'apple'",
			shouldAlert: true,
		},

		"UPDATE using indexed name": {
			query:       "UPDATE vegetables SET created_at = NOW() WHERE name = 'potato'",
			shouldAlert: false,
		},

		"UPDATE using non-indexed name": {
			query:       "UPDATE fruits SET created_at = NOW() WHERE name = 'apple'",
			shouldAlert: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			r := require.New(t)
			db := testenv.InitPgDB(t)

			var alerted bool
			var explainResult string

			options := scanalert.DefaultAlertOptions()
			options.QueryType = scanalert.RawQuery

			scanalert.RegisterScanAlert(db, options, func(source string, result string) {
				explainResult = result
				alerted = true
			})

			err := db.Exec(tc.query).Error
			r.NoError(err)

			if tc.shouldAlert {
				r.True(alerted, "should have alerted:\n%s", explainResult)
			} else {
				r.False(alerted, "should not have alerted\n%s", explainResult)
			}

		})
	}
}
