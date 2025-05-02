package pool

import (
	"http_grpc/internal/repository/model"
	"sync"
)

// TaskData 任务处理数据
type TaskData struct {
	UserData model.User
}

// Reset 重置 TaskData
func (data *TaskData) Reset() {
	data.UserData = model.User{}
}

// TaskDataPool 任务数据对象池
var TaskDataPool = sync.Pool{
	New: func() interface{} {
		return new(TaskData)
	},
}
