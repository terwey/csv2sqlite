package stats

import "testing"

func TestStatistics_Add(t *testing.T) {
	type Col struct {
		ColumnInfo
		Count int
	}
	tests := []struct {
		name  string
		names []string
		cols  []*Col
	}{
		{
			name:  "two cols",
			names: []string{"foo", "bar"},
			cols: []*Col{
				{
					ColumnInfo: ColumnInfo{
						Length: 5,
						Pos:    0, // foo
					},
					Count: 1,
				},
				{
					ColumnInfo: ColumnInfo{
						Length: 10,
						Pos:    1, // bar
					},
					Count: 1,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewStatistics(tt.names)
			for _, col := range tt.cols {
				s.Add(&col.ColumnInfo)
			}

			s.Rows = 1
			s.Calculate()

			for _, ci := range tt.cols {
				if *s.data[ci.Pos].LengthDistribution[length(ci.Length)] != uint64(ci.Count) {
					t.Errorf("Length for pos does not match: %d vs %d", *s.data[ci.Pos].LengthDistribution[length(ci.Length)], ci.Length)
				}
			}
		})
	}
}

func TestStatisticMerge(t *testing.T) {
	names := []string{"foo", "bar"}
	s1 := NewStatistics(names)

	// s1 contains 15 for pos 0 length
	s1.Add(&ColumnInfo{5, ColumnType_Text, 0})
	s1.Add(&ColumnInfo{10, ColumnType_Text, 0})

	// s1 contains 50 for pos 1 length
	s1.Add(&ColumnInfo{20, ColumnType_Text, 1})
	s1.Add(&ColumnInfo{30, ColumnType_Text, 1})

	if s1.data[0].AverageLengthSum != 15 {
		t.Errorf("should contain 15 on pos 0")
	}

	if s1.data[1].AverageLengthSum != 50 {
		t.Errorf("should contain 50 on pos 1")
	}

	s2 := NewStatistics(names)
	// s2 contains 8 for pos 0 length
	s2.Add(&ColumnInfo{2, ColumnType_Text, 0})
	s2.Add(&ColumnInfo{6, ColumnType_Text, 0})

	// s2 contains 7 for pos 1 length
	s2.Add(&ColumnInfo{3, ColumnType_Text, 1})
	s2.Add(&ColumnInfo{4, ColumnType_Text, 1})

	if s2.data[0].AverageLengthSum != 8 {
		t.Errorf("should contain 8 on pos 0")
	}

	if s2.data[1].AverageLengthSum != 7 {
		t.Errorf("should contain 7 on pos 1")
	}

	s1.Merge(s2)

	if s1.data[0].AverageLengthSum != 23 {
		t.Errorf("should contain 23 on pos 0")
	}

	if s1.data[1].AverageLengthSum != 57 {
		t.Errorf("should contain 57 on pos 1")
	}

	// merge it again but twice
	s1.Merge(s2, s2)

	if s1.data[0].AverageLengthSum != 39 {
		t.Errorf("should contain 39 on pos 0")
	}

	if s1.data[1].AverageLengthSum != 71 {
		t.Errorf("should contain 71 on pos 1")
	}
}
