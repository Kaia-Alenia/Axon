package main
type InputSimulator interface {
	MoveMouse(dx, dy float64)
	Click(button string)
	MouseDown(button string)
	MouseUp(button string)
	Scroll(direction string)
	Type(text string)
	Key(key string)
	KeyCombo(modifier, key string)
	Close()
}
func NewSimulator() InputSimulator {
	return initSimulator()
}
