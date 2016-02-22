package main

type Tag struct {
	Name        string
	Attributes  map[string]string
	TextContent string
	Children    []*Tag
	Parent      *Tag
}

func (self *Tag) FirstChild(name string) *Tag {
	for _, child := range self.Children {
		if child.Name == name {
			return child
		}
	}
	return nil
}
