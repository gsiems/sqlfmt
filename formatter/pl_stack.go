package formatter

type plStack struct {
	stk []string
}

func (s *plStack) Indents() int {
	stackIndent := 0
	for _, v := range s.stk {
		switch v {
		case "CASE", "EXCEPTION":
			stackIndent += 2
		default:
			stackIndent += 1
		}
	}
	return stackIndent
}

func (s *plStack) IsEmpty() bool {
	return len(s.stk) < 1
}

func (s *plStack) Last() (v string) {
	iMax := s.Length() - 1
	if iMax >= 0 {
		v = s.stk[iMax]
	}
	return v
}

func (s *plStack) LastBlock() (v string) {
	iMax := s.Length() - 1
	if iMax >= 0 {
		for idx := iMax; idx >= 0; idx-- {
			switch s.stk[idx] {
			case "DECLARE", "BEGIN", "EXCEPTION":
				return s.stk[idx]
			}
		}
	}
	return ""
}

func (s *plStack) Length() int {
	return len(s.stk)
}

func (s *plStack) Pop() (v string) {

	v = s.Last()

	iMax := s.Length() - 1
	if iMax > 0 {
		s.stk = s.stk[:iMax]
	} else {
		s.Reset()
	}

	return v
}

func (s *plStack) Push(v string) {

	switch s.Last() {
	case "DECLARE":
	// nada. only need one declare
	default:
		s.stk = append(s.stk, v)
	}

}

func (s *plStack) Reset() {
	s.stk = make([]string, 0)
}

func (s *plStack) Set(v string) {
	iMax := s.Length() - 1
	if iMax >= 0 {
		s.stk[iMax] = v
	} else {
		s.Push(v)
	}
}

func (s *plStack) Upsert(v string) {

	switch v {
	case "FUNCTION", "PACKAGE", "PROCEDURE", "TRIGGER", "IS", "AS", "DECLARE":
		s.Push(v)

	case "BEGIN":
		if s.Length() > 0 {
			switch s.Last() {
			case "FUNCTION", "PACKAGE", "PROCEDURE", "TRIGGER", "IS", "AS", "DECLARE":
				s.Set(v)
			default:
				s.Push(v)
			}
		} else {
			s.Push(v)
		}
	case "EXCEPTION":
		if s.Length() > 0 {
			switch s.Last() {
			case "BEGIN":
				s.Set(v)
			}
		}
	}
}
