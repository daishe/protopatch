package protopatch

import "google.golang.org/protobuf/proto"

func Move(base proto.Message, targetPath, replacementPath string, opts ...Option) error {
	return moveWithSetup(base, targetPath, replacementPath, newSetup(opts...))
}

func moveWithSetup(base proto.Message, targetPath, replacementPath string, setup *setup) error {
	if targetPath == replacementPath { // set value pointed by path to itself
		return setToItself(base, targetPath, setup)
	}

	replacementValue, replacementSet, err := getCopyAndSetter(base, replacementPath, setup)
	if err != nil {
		return err
	}
	targetValue, targetSet, err := getCopyAndSetter(base, targetPath, setup) // TODO: Create and use getAndSetter function to avoid copying target
	if err != nil {
		return err
	}

	err = targetSet(replacementValue)
	if err != nil {
		return err
	}
	err = replacementSet(nil)
	if err != nil {
		_ = targetSet(targetValue) // attempt target restore; ignore errors
		return err
	}
	return nil
}

// func moveWithSetup(base proto.Message, targetPath, replacementPath string, setup *setup) error {
// 	if targetPath == replacementPath { // set value pointed by path to itself
// 		return setToItself(base, targetPath, setup)
// 	}

// 	c := MessageContainer(base)

// 	var replacementValue any
// 	var restore func()
// 	if replacementPath == "" { // replacement path refers to base message
// 		replacementValue = proto.Clone(base)
// 		err := c.(*messageContainer).setSelf(nil)
// 		if err != nil {
// 			return err
// 		}
// 		restore = func() {
// 			_ = c.(*messageContainer).setSelf(replacementValue) // ignore errors
// 		}
// 	} else if last := Path(replacementPath).Last(); !last.IsFirst() { // replacement path has more than 1 element
// 		replacementContainer, err := access(c, last.PrecedingPath(), setup)
// 		if err != nil {
// 			return err
// 		}
// 		replacementValue, err = replacementContainer.GetCopy(last.Value())
// 		if err != nil {
// 			return NewErrInPath(string(last.PrecedingPath()), err)
// 		}
// 		err = replacementContainer.Set(last.Value(), nil)
// 		if err != nil {
// 			return NewErrInPath(string(last.PrecedingPath()), err)
// 		}
// 		restore = func() {
// 			_ = replacementContainer.Set(last.Value(), replacementValue) // ignore errors
// 		}
// 	} else { // path has only 1 element
// 		replacementContainer, err := transformContainer(c, setup)
// 		if err != nil {
// 			return err
// 		}
// 		replacementValue, err = replacementContainer.GetCopy(replacementPath)
// 		if err != nil {
// 			return err
// 		}
// 		err = replacementContainer.Set(replacementPath, nil)
// 		if err != nil {
// 			return err
// 		}
// 		restore = func() {
// 			_ = replacementContainer.Set(last.Value(), replacementValue) // ignore errors
// 		}
// 	}

// 	attemptRestore := func() {
// 		if restore != nil {
// 			restore()
// 		}
// 	}
// 	defer attemptRestore()

// 	if targetPath == "" { // target path refers to base message
// 		err := setSelf(base, replacementValue, setup)
// 		if err != nil {
// 			return err
// 		}
// 	} else if last := Path(targetPath).Last(); !last.IsFirst() { // target path has more than 1 element
// 		targetContainer, err := access(c, last.PrecedingPath(), setup)
// 		if err != nil {
// 			return err
// 		}
// 		ref, err := targetContainer.GetNew(last.Value())
// 		if err != nil {
// 			return NewErrInPath(string(last.PrecedingPath()), err)
// 		}
// 		conv, err := convert(ref, replacementValue, setup)
// 		if err != nil {
// 			return NewErrInPath(string(last.PrecedingPathWithCurrentSegment()), err)
// 		}
// 		err = targetContainer.Set(last.Value(), conv)
// 		if err != nil {
// 			return NewErrInPath(string(last.PrecedingPath()), err)
// 		}
// 	} else { // path has only 1 element
// 		targetContainer, err := transformContainer(c, setup)
// 		if err != nil {
// 			return err
// 		}
// 		ref, err := targetContainer.GetNew(targetPath)
// 		if err != nil {
// 			return err
// 		}
// 		conv, err := convert(ref, replacementValue, setup)
// 		if err != nil {
// 			return NewErrInPath(targetPath, err)
// 		}
// 		err = targetContainer.Set(last.Value(), conv)
// 		if err != nil {
// 			return err
// 		}
// 	}
// 	restore = nil
// 	return nil
// }
