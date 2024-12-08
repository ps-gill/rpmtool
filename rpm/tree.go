package rpm

type Tree struct {
	BuildDir, RpmDir, SourceDir, SpecDir, SrpmDir string
}

func GetTree() (*Tree, error) {
	buildDir, err := ExpandMacro(macroBuildDir)
	if err != nil {
		return nil, err
	}

	rpmDir, err := ExpandMacro(macroRpmDir)
	if err != nil {
		return nil, err
	}

	sourceDir, err := ExpandMacro(macroSourceDir)
	if err != nil {
		return nil, err
	}

	specDir, err := ExpandMacro(macroSpecDir)
	if err != nil {
		return nil, err
	}

	srpmDir, err := ExpandMacro(macroSrpmDir)
	if err != nil {
		return nil, err
	}

	return &Tree{
		BuildDir:  buildDir,
		RpmDir:    rpmDir,
		SourceDir: sourceDir,
		SpecDir:   specDir,
		SrpmDir:   srpmDir,
	}, nil
}
