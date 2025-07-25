package erz

import (
	"runtime"
	"strings"
)

type StackFrame struct {
	Function string `json:"function"`
	File     string `json:"file"`
	Line     int    `json:"line"`
}

func captureStackTrace(skip int) []StackFrame {
	var frames []StackFrame

	for i := skip; i < skip+10; i++ {
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}

		fn := runtime.FuncForPC(pc)
		var funcName string
		if fn != nil {
			funcName = fn.Name()
			if idx := strings.LastIndex(funcName, "/"); idx != -1 {
				funcName = funcName[idx+1:]
			}
		}

		if idx := strings.LastIndex(file, "/"); idx != -1 {
			file = file[idx+1:]
		}

		frames = append(
			frames, StackFrame{
				Function: funcName,
				File:     file,
				Line:     line,
			},
		)
	}

	return frames
}
