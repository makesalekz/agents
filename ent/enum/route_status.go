package enum

type RouteStatus string

const (
	Planned    RouteStatus = "PLANNED"
	InProgress RouteStatus = "IN_PROGRESS"
	Completed  RouteStatus = "COMPLETED"
)

func (r RouteStatus) Value() string {
	return string(r)
}

func (r RouteStatus) IsValid() bool {
	switch r {
	case Planned, InProgress, Completed:
		return true
	}
	return false
}

func (RouteStatus) Values() []string {
	return []string{
		Planned.Value(),
		InProgress.Value(),
		Completed.Value(),
	}
}
