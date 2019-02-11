package graph

//func TestParse(t *testing.T) {
//	it := NewTaskInstanceIterator(map[string]*types.TaskInvocation{
//		"a": {
//			Task: types.NewTask("a", "noop"),
//		},
//		"b": {
//			Task: &types.Task{
//				Metadata: types.NewObjectMetadata("b"),
//				Spec: &types.TaskSpec{
//					Requires: map[string]*types.TaskDependencyParameters{
//						"a": nil,
//					},
//				},
//			},
//		},
//		"c": {
//			Task: &types.Task{
//				Metadata: types.NewObjectMetadata("c"),
//				Spec: &types.TaskSpec{
//					Requires: map[string]*types.TaskDependencyParameters{
//						"a": nil,
//						"b": nil,
//					},
//				},
//			},
//		},
//	})
//	dag := Parse(it)
//	root := Get(dag, "a")
//	assert.NotNil(t, root)
//
//	assert.Equal(t, 2, len(dag.From(root)))
//	assert.Equal(t, 0, len(dag.To(root)))
//	nodeB := Get(dag, "b")
//	assert.NotNil(t, nodeB)
//	assert.Equal(t, 1, len(dag.From(nodeB)))
//	assert.Equal(t, 1, len(dag.To(nodeB)))
//	nodeC := Get(dag, "c")
//	assert.NotNil(t, nodeC)
//	assert.Equal(t, 0, len(dag.From(nodeC)))
//	assert.Equal(t, 2, len(dag.To(nodeC)))
//}
//
//func TestParseDynamic(t *testing.T) {
//	it := NewTaskInstanceIterator(map[string]*types.TaskInstance{
//		"a": {
//			Task: types.NewTask("a", "noop"),
//		},
//		"b": {
//			Task: &types.Task{
//				Metadata: types.NewObjectMetadata("b"),
//				Spec: &types.TaskSpec{
//					Requires: map[string]*types.TaskDependencyParameters{
//						"a": nil,
//					},
//				},
//			},
//		},
//		"c": {
//			Task: &types.Task{
//				Metadata: types.NewObjectMetadata("c"),
//				Spec: &types.TaskSpec{
//					Requires: map[string]*types.TaskDependencyParameters{
//						"a": nil,
//						"b": nil,
//					},
//				},
//			},
//		},
//		"d": {
//			Task: &types.Task{
//				Metadata: types.NewObjectMetadata("d"),
//				Spec: &types.TaskSpec{
//					Requires: map[string]*types.TaskDependencyParameters{
//						"b": {
//							Type: types.TaskDependencyParameters_DYNAMIC_OUTPUT,
//						},
//					},
//				},
//			},
//		},
//	})
//
//	dag := Parse(it)
//	root := Get(dag, "a")
//	assert.NotNil(t, root)
//
//	assert.Equal(t, 2, len(dag.From(root)))
//	assert.Equal(t, 0, len(dag.To(root)))
//	nodeB := Get(dag, "b")
//	assert.NotNil(t, nodeB)
//	assert.Equal(t, 2, len(dag.From(nodeB)))
//	assert.Equal(t, 1, len(dag.To(nodeB)))
//	nodeC := Get(dag, "c")
//	assert.NotNil(t, nodeC)
//	assert.Equal(t, 0, len(dag.From(nodeC)))
//	assert.Equal(t, 3, len(dag.To(nodeC)))
//	nodeD := Get(dag, "d")
//	assert.NotNil(t, nodeD)
//	assert.Equal(t, 1, len(dag.From(nodeD)))
//	assert.Equal(t, 1, len(dag.To(nodeD)))
//}

// TODO FIX
