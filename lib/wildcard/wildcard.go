package wildcard

const (
	normal      = iota
	all         //*
	any         //?
	setSymbol   //[]
	rangeSymbol //[a-z]
	negSymbol   //[^a]
)

type item struct {
	character byte
	set       map[byte]bool
	typeCode  int
}

func (i *item) contains(c byte) bool {
	if i.typeCode == setSymbol { //set type
		_, ok := i.set[c]
		return ok
	} else if i.typeCode == rangeSymbol { //range type
		if _, ok := i.set[c]; ok {
			return true
		}
		var (
			min uint8 = 255
			max uint8 = 0
		)
		for k := range i.set {
			if k < min {
				min = k
			}
			if k > max {
				max = k
			}
		}
		return c >= min && c <= max
	} else { //normal type
		_, ok := i.set[c]
		return !ok
	}
}

type Pattern struct {
	items []*item
}

// CompilePattern compiles a wildcard pattern into a Pattern struct.
func CompilePattern(src string) *Pattern {
	items := make([]*item, 0)
	escape := false       // escape flag
	inSet := false        // in set flag
	var set map[byte]bool // set for set type
	for _, c := range src {
		s := byte(c)
		if escape { // if the last character was a backslash, treat this character as normal
			items = append(items, &item{character: s, typeCode: normal})
			escape = false
		} else if s == '*' {
			items = append(items, &item{typeCode: all})
		} else if s == '?' {
			items = append(items, &item{typeCode: any})
		} else if s == '\\' {
			escape = true // set escape flag
		} else if s == '[' {
			if !inSet {
				inSet = true              // set in set flag
				set = make(map[byte]bool) // create a new set
			} else {
				set[s] = true // add the character to the set
			}
		} else if s == ']' {
			if inSet {
				inSet = false // unset in set flag
				typeCode := setSymbol
				if _, ok := set['-']; ok { // if the set contains ^, set typeCode to negSymbol
					typeCode = negSymbol
					delete(set, '-')
				}
				if _, ok := set['^']; ok { // if the set contains ^, set typeCode to negSymbol
					typeCode = negSymbol
					delete(set, '^')
				}
				items = append(items, &item{set: set, typeCode: typeCode}) // add the set to the items
			} else { //common case, treat as normal character
				items = append(items, &item{character: s, typeCode: normal}) // add normal character to items
			}
		} else {
			if inSet { // if in set, add the character to the set
				set[s] = true // add the character to the set
			} else { // common case, treat as normal character
				items = append(items, &item{character: s, typeCode: normal}) // add normal character to items
			}
		}
	}
	return &Pattern{
		items: items,
	}
}

func (p *Pattern) Match(src string) bool {
	if len(p.items) == 0 {
		return len(src) == 0
	}
	m:=len(src)
	n:=len(p.items)
	table := make([][]bool, m+1) // create a table to store the match results
	for i := 0; i <= m; i++ {
		table[i] = make([]bool, n+1) // create a new row in the table
	}
	table[0][0] = true // empty pattern matches empty string
	for j := 1; j <= n; j++ { // fill the first row of the table
		table[0][j] = table[0][j-1] && p.items[j-1].typeCode == all // if the previous item is all, set the current item to true
	}
	for i := 1; i <= m; i++ { // fill the rest of the table
		for j := 1; j <= n; j++ { // fill the current row of the table
			if p.items[j-1].typeCode == all { // if the current item is all, set the current cell to true if the previous cell is true
				table[i][j] = table[i][j-1] || table[i-1][j]
			}else{
				table[i][j] = table[i-1][j-1] &&
				(p.items[j-1].typeCode == any ||
					(p.items[j-1].typeCode == normal && uint8(src[i-1]) == p.items[j-1].character) ||
					(p.items[j-1].typeCode >= setSymbol && p.items[j-1].contains(src[i-1])))
			}

		}
	}
	return table[m][n] // return the result of the last cell in the table
}
