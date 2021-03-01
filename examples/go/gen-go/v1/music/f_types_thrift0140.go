// +build thrift14

package music

import (
	"context"
	"fmt"

	"github.com/apache/thrift/lib/go/thrift"
)

func (p *Track) Read(ctx context.Context, iprot thrift.TProtocol) error {
	if _, err := iprot.ReadStructBegin(ctx); err != nil {
		return thrift.PrependError(fmt.Sprintf("%T read error: ", p), err)
	}

	for {
		_, fieldTypeId, fieldId, err := iprot.ReadFieldBegin(ctx)
		if err != nil {
			return thrift.PrependError(fmt.Sprintf("%T field %d read error: ", p, fieldId), err)
		}
		if fieldTypeId == thrift.STOP {
			break
		}
		switch fieldId {
		case 1:
			if err := p.ReadField1(ctx, iprot); err != nil {
				return err
			}
		case 2:
			if err := p.ReadField2(ctx, iprot); err != nil {
				return err
			}
		case 3:
			if err := p.ReadField3(ctx, iprot); err != nil {
				return err
			}
		case 4:
			if err := p.ReadField4(ctx, iprot); err != nil {
				return err
			}
		case 5:
			if err := p.ReadField5(ctx, iprot); err != nil {
				return err
			}
		case 6:
			if err := p.ReadField6(ctx, iprot); err != nil {
				return err
			}
		default:
			if err := iprot.Skip(ctx, fieldTypeId); err != nil {
				return err
			}
		}
		if err := iprot.ReadFieldEnd(ctx); err != nil {
			return err
		}
	}
	if err := iprot.ReadStructEnd(ctx); err != nil {
		return thrift.PrependError(fmt.Sprintf("%T read struct end error: ", p), err)
	}
	return nil
}

func (p *Track) ReadField1(ctx context.Context, iprot thrift.TProtocol) error {
	if v, err := iprot.ReadString(ctx); err != nil {
		return thrift.PrependError("error reading field 1: ", err)
	} else {
		p.Title = v
	}
	return nil
}

func (p *Track) ReadField2(ctx context.Context, iprot thrift.TProtocol) error {
	if v, err := iprot.ReadString(ctx); err != nil {
		return thrift.PrependError("error reading field 2: ", err)
	} else {
		p.Artist = v
	}
	return nil
}

func (p *Track) ReadField3(ctx context.Context, iprot thrift.TProtocol) error {
	if v, err := iprot.ReadString(ctx); err != nil {
		return thrift.PrependError("error reading field 3: ", err)
	} else {
		p.Publisher = v
	}
	return nil
}

func (p *Track) ReadField4(ctx context.Context, iprot thrift.TProtocol) error {
	if v, err := iprot.ReadString(ctx); err != nil {
		return thrift.PrependError("error reading field 4: ", err)
	} else {
		p.Composer = v
	}
	return nil
}

func (p *Track) ReadField5(ctx context.Context, iprot thrift.TProtocol) error {
	if v, err := iprot.ReadDouble(ctx); err != nil {
		return thrift.PrependError("error reading field 5: ", err)
	} else {
		temp := Minutes(v)
		p.Duration = temp
	}
	return nil
}

func (p *Track) ReadField6(ctx context.Context, iprot thrift.TProtocol) error {
	if v, err := iprot.ReadI32(ctx); err != nil {
		return thrift.PrependError("error reading field 6: ", err)
	} else {
		temp := PerfRightsOrg(v)
		p.Pro = temp
	}
	return nil
}
