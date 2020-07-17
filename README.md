# comtrade
Go support for IEEE COMTRADE readers

### usage

```
  cfg := new(CFG)
	c, err := ioutil.ReadFile("files/ANGLEDS.CFG")
	if err != nil {
		t.Error(err)
	}
	err = cfg.UnmarshalCfg(c)
	if err != nil {
		t.Error(err)
	}

	dat, err := ioutil.ReadFile("files/ANGLEDS.DAT")
	if err != nil {
		t.Error(err)
	}
	result, err := cfg.UnmarshalDat(dat)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(result)
```
