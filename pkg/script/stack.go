package script

type Stack struct {
	items []string
}

func NewStack() *Stack {
	return &Stack{}
}

func (s *Stack) Push(item string) {
	s.items = append(s.items, item)
}

func (s *Stack) Pop() string {
	item := s.items[len(s.items)-1]
	s.items = s.items[:len(s.items)-1]
	return item
}

func (s *Stack) Len() uint {
	return uint(len(s.items))
}
