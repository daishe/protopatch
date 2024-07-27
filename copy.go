package protopatch

import "google.golang.org/protobuf/proto"

func Copy(base proto.Message, targetPath, replacementPath string, opts ...Option) error {
	return copyWithSetup(base, targetPath, replacementPath, newSetup(opts...))
}

func copyWithSetup(base proto.Message, targetPath, replacementPath string, setup *setup) error {
	if targetPath == replacementPath { // set value pointed by path to itself
		return setToItself(base, targetPath, setup)
	}

	var replacementValue any
	if replacementPath == "" { // replacement path refers to base message
		replacementValue = proto.Clone(base)
	} else if last := Path(replacementPath).Last(); !last.IsFirst() { // replacement path has more than 1 element
		replacementContainer, err := access(MessageContainer(base), last.PrecedingPath(), setup)
		if err != nil {
			return err
		}
		replacementValue, err = replacementContainer.GetCopy(last.Value())
		if err != nil {
			return NewErrInPath(string(last.PrecedingPath()), err)
		}
	} else { // path has only 1 element
		replacementContainer, err := transformContainer(MessageContainer(base), setup)
		if err != nil {
			return err
		}
		replacementValue, err = replacementContainer.GetCopy(replacementPath)
		if err != nil {
			return err
		}
	}

	return setWithSetup(base, targetPath, replacementValue, setup)
}
