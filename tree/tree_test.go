package tree

import "testing"

func TestTreeFlatten(t *testing.T) {
	var rootNode = TreeNode{
		ID:    "A",
		Depth: 0,
		children: []*TreeNode{
			&TreeNode{
				ID:    "B",
				Depth: 1,
				children: []*TreeNode{
					&TreeNode{ID: "C", Depth: 2},
					&TreeNode{ID: "D", Depth: 2},
				},
			},
			&TreeNode{
				ID:    "E",
				Depth: 1,
				children: []*TreeNode{
					&TreeNode{ID: "F", Depth: 2},
					&TreeNode{ID: "G",
						Depth:    2,
						children: []*TreeNode{&TreeNode{ID: "H", Depth: 3}}},
				},
			},
		},
	}

	// Act
	var flat = rootNode.Flatten()

	// Assert
	var value, exists = flat["H"]
	if !exists {
		t.Error("Looking for key `H` in the flattened tree returned no value.")
		t.FailNow()
	} else if value != 3 {
		t.Errorf("Depth of key `H` should be 3 but was %v", value)
		t.FailNow()
	}
}
