package pb

import (
	"strconv"

	"github.com/pipe01/flydigictl/pkg/flydigi/config"
	"golang.org/x/image/colornames"
)

func ColorFromRGB(r, g, b byte) *Color {
	return &Color{Rgb: int32(b) | (int32(g) << 8) | (int32(r) << 16)}
}

func ColorFromHex(hex string) (*Color, bool) {
	if hex[0] == '#' {
		hex = hex[1:]
	}
	if len(hex) != 6 {
		return nil, false
	}

	n, err := strconv.ParseUint(hex, 16, 32)
	if err != nil || n > 0xFFFFFF {
		return nil, false
	}

	return &Color{Rgb: int32(n)}, true
}

func ColorFromName(name string) (*Color, bool) {
	rgba, ok := colornames.Map[name]
	if !ok {
		return nil, false
	}

	return ColorFromRGB(rgba.R, rgba.G, rgba.B), true
}

func (c *Color) RGB() (r, g, b byte) {
	return byte(c.Rgb >> 16), byte(c.Rgb >> 8), byte(c.Rgb)
}

func (c *Color) LedUnit() *config.LedUnit {
	r, g, b := c.RGB()
	return &config.LedUnit{r, g, b}
}

func ConvertGamepadConfiguration(bean *config.AllConfigBean) *GamepadConfiguration {
	return &GamepadConfiguration{
		LeftJoystick:  ConvertJoystickConfiguration(bean.JoyMapping.LeftJoystic),
		RightJoystick: ConvertJoystickConfiguration(bean.JoyMapping.RightJoystic),
	}
}

func (c *GamepadConfiguration) ApplyTo(bean *config.AllConfigBean) {
	c.LeftJoystick.ApplyTo(bean.JoyMapping.LeftJoystic)
	c.RightJoystick.ApplyTo(bean.JoyMapping.RightJoystic)
}

func ConvertJoystickConfiguration(bean *config.JoyStickBean) *JoystickConfiguration {
	return &JoystickConfiguration{
		Deadzone: bean.Curve.Zero,
	}
}

func (c *JoystickConfiguration) ApplyTo(bean *config.JoyStickBean) {
	bean.Curve.Zero = c.Deadzone
}

func ConvertLEDConfiguration(bean *config.NewLedConfigBean) *LedsConfiguration {
	var leds isLedsConfiguration_Leds
	colorAt := func(group, unit int) *Color {
		if len(bean.LedGroups) <= group || len(bean.LedGroups[group].Units) <= unit {
			return ColorFromRGB(0, 0, 0)
		}

		u := bean.LedGroups[group].Units[unit]
		return ColorFromRGB(u.R, u.G, u.B)
	}

	switch bean.LedMode {
	case config.LedModeOff:
		leds = &LedsConfiguration_Off{}

	case config.LedModeSteady:
		unit := bean.LedGroups[0].Units[0]

		leds = &LedsConfiguration_Steady{
			Steady: &LedsSteady{
				Color: ColorFromRGB(unit.R, unit.G, unit.B),
			},
		}

	case config.LedModeStreamlined:
		leds = &LedsConfiguration_Streamlined{
			Streamlined: &LedsStreamlined{
				Speed: float32(100-bean.Loop_time) / 100,
			},
		}

	case config.LedModeBreathing:
		leds = &LedsConfiguration_Breathing{
			Breathing: &LedsBreathing{
				Color: colorAt(0, 0),
				Speed: float32(100-bean.Loop_time) / 100,
			},
		}

	case config.LedModeGradient:
		leds = &LedsConfiguration_Gradient{
			Gradient: &LedsGradient{
				StartColor: colorAt(0, 0),
				EndColor:   colorAt(1, 1),
				Speed:      float32(100-bean.Loop_time) / 100,
			},
		}

	case config.LedModeFeedback:
		leds = &LedsConfiguration_Feedback{
			Feedback: &LedsFeedback{
				Speed: float32(100-bean.Loop_time) / 100,
			},
		}

	default:
		leds = nil
	}

	return &LedsConfiguration{
		Leds:       leds,
		Brightness: float32(bean.Light_scale) / 255,
	}
}

func (c *LedsConfiguration) ApplyTo(bean *config.NewLedConfigBean) {
	bean.Light_scale = byte(c.Brightness * 100)

	switch leds := c.Leds.(type) {
	case *LedsConfiguration_Off:
		bean.SetOff()

	case *LedsConfiguration_Steady:
		bean.SetSteady(*leds.Steady.Color.LedUnit())

	case *LedsConfiguration_Streamlined:
		bean.SetStreamlined(leds.Streamlined.Speed)

	case *LedsConfiguration_Breathing:
		bean.SetBreathing(*leds.Breathing.Color.LedUnit(), leds.Breathing.Speed)

	case *LedsConfiguration_Gradient:
		bean.SetGradient(*leds.Gradient.StartColor.LedUnit(), *leds.Gradient.EndColor.LedUnit(), leds.Gradient.Speed)

	case *LedsConfiguration_Feedback:
		bean.SetFeedback(leds.Feedback.Speed)
	}
}
