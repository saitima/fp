package xxx

func Add4(c, a, b FieldElement) {
	add4(c.(*Fe256), a.(*Fe256), b.(*Fe256))
}
func add4(a, b, c *Fe256)
