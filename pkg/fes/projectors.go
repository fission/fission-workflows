package fes

var DefaultProjector = SimpleProjector{}

func Project(target Aggregator, events ...*Event) error {
	return DefaultProjector.Project(target, events...)
}

type SimpleProjector struct{}

func (rp *SimpleProjector) Project(target Aggregator, events ...*Event) error {
	for _, event := range events {
		err := rp.project(target, event)
		if err != nil {
			return err
		}
	}

	return nil
}

func (rp *SimpleProjector) project(target Aggregator, event *Event) error {
	return target.ApplyEvent(event)
}
