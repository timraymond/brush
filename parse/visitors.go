package parse

// A CompositeVisitor takes an arbitrary number of Visitors and invokes their
// corresponding Accept methods when its Accept methods are invoked. This
// allows for one traversal of a Brush AST with an arbitrary number of Visitors.
type CompositeVisitor struct {
	visitors []Visitor
}

// Takes a variadic list of Visitors to allow for easy creation of
// CompositeVisitors.
func NewCompositeVisitor(visitors ...Visitor) *CompositeVisitor {
	return &CompositeVisitor{visitors}
}

// Accepts BraaiTagNodes and dispatches to the corresponding AcceptTag method
// of the internal list of visitors.
func (cv *CompositeVisitor) AcceptTag(b *BraaiTagNode) {
	for _, visitor := range cv.visitors {
		visitor.AcceptTag(b)
	}
}

// Accepts BlockTagNodes and dispatches to the corresponding AcceptBlockTag method
// of the internal list of visitors.
func (cv *CompositeVisitor) AcceptBlockTag(b *BlockTagNode) {
	for _, visitor := range cv.visitors {
		visitor.AcceptBlockTag(b)
	}
}

// Accepts TextNodes and dispatches to the corresponding AcceptTextNode method
// of the internal list of visitors.
func (cv *CompositeVisitor) AcceptTextNode(b *TextNode) {
	for _, visitor := range cv.visitors {
		visitor.AcceptTextNode(b)
	}
}
