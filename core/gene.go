package core

type Gene struct {
	InputSize  int
	OutputSize int
	Weights    [][]float64 // [OutputSize][InputSize]
	Biases     []float64   // [OutputSize]
}

func NewGene(inputSize, outputSize int) Gene {
	if inputSize <= 0 || outputSize <= 0 {
		return Gene{}
	}

	weights := make([][]float64, outputSize)
	for i := range weights {
		weights[i] = make([]float64, inputSize)
	}

	biases := make([]float64, outputSize)

	return Gene{
		InputSize:  inputSize,
		OutputSize: outputSize,
		Weights:    weights,
		Biases:     biases,
	}
}

// Compute 前向计算：根据输入计算输出
func (g *Gene) Compute(inputs []float64) []float64 {
	if g == nil {
		return nil
	}

	return inputs // 这里先直接返回输入，后续再实现真正的前向计算
}

// Adjust 根据 REINFORCE 算法微调权重和偏置
// inputs: 当前输入  outputIdx: 要调整的输出索引 reward: 本次调整的奖励（正数或负数） lr: 学习率（调整幅度）
func (g *Gene) Adjust(inputs []float64, outputIdx int, reward, lr float64) {
	if g == nil {
		return
	}
	if outputIdx < 0 || outputIdx >= g.OutputSize {
		return
	}

	// 这里先不实现真正的调整逻辑，后续再根据 REINFORCE 算法微调权重和偏置
}

// Mutate 根据给定的变异率随机调整权重和偏置
func (g *Gene) Mutate(rate float64) {
	if g == nil {
		return
	}
	if rate < 0 || rate > 1 {
		return
	}

	// 这里先不实现真正的变异逻辑，后续再根据变异率随机调整权重和偏置
}
