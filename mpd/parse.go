package mpd

import (
	"strconv"

	"github.com/pkg/errors"

	"github.com/fhs/gompd/mpd"
)

type parseMap struct {
	Err error

	m mpd.Attrs
}

func (p *parseMap) Float(field string, target *float64, optional bool) bool {
	*target = 0 // Resets the target
	if p.Err != nil {
		return false
	}

	var err error
	if val, ok := p.m[field]; ok {
		if *target, err = strconv.ParseFloat(val, 64); err != nil {
			err = errors.Wrapf(err, "Field `%s` = `%s` parsing failed", field, val)
		}
	} else {
		err = errors.Errorf("Field `%s` not empty", field)
	}

	if !optional {
		p.Err = err
	}

	return err == nil
}

func (p *parseMap) String(field string, target *string, optional bool) bool {
	*target = "" // Resets the target
	if p.Err != nil {
		return false
	}

	var err error
	if val, ok := p.m[field]; ok {
		*target = val
	} else {
		err = errors.Errorf("Field `%s` not empty", field)
	}

	if !optional {
		p.Err = err
	}

	return err == nil
}

func (p *parseMap) Int(field string, target *int, optional bool) bool {
	*target = 0 // Resets the target
	if p.Err != nil {
		return false
	}

	var err error
	if val, ok := p.m[field]; ok {
		if *target, err = strconv.Atoi(val); err != nil {
			err = errors.Wrapf(err, "Field `%s` = `%s` parsing failed", field, val)
		}
	} else {
		err = errors.Errorf("Field `%s` not empty", field)
	}

	if !optional {
		p.Err = err
	}

	return err == nil
}

func (p *parseMap) Bool(field string, target *bool, optional bool) bool {
	*target = false // Resets the target
	if p.Err != nil {
		return false
	}

	var err error
	if val, ok := p.m[field]; ok {
		switch val {
		case "0":
			*target = false
		case "1":
			*target = true
		default:
			err = errors.Errorf("Field `%s` = `%s` parsing failed: expected 0 or 1", field, val)
		}
	} else {
		err = errors.Errorf("Field `%s` not empty", field)
	}

	if !optional {
		p.Err = err
	}

	return err == nil
}
