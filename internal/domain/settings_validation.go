package domain

import "time"

const DateDisplayLocaleDefault = "locale_default"

type Settings struct {
	Language                  LanguageCode
	Locale                    string
	WeekStartDay              WeekStartDay
	TimeZone                  string
	DateDisplayFormat         string
	TimeDisplayFormat         TimeDisplayFormat
	DefaultTimeZoneMode       DefaultTimeZoneMode
	TodayPrimaryLookaheadDays int
	DeadlineAwarenessDays     int
	WeatherCity               string
}

func DefaultSettings(language LanguageCode, deviceTimeZone string) Settings {
	if !IsValidLanguageCode(language) {
		language = LanguageEnglish
	}
	if deviceTimeZone == "" {
		deviceTimeZone = "UTC"
	}

	settings := Settings{
		Language:                  language,
		Locale:                    DefaultLocale(language),
		WeekStartDay:              DefaultWeekStartDay(language),
		TimeZone:                  deviceTimeZone,
		DateDisplayFormat:         DateDisplayLocaleDefault,
		TimeDisplayFormat:         DefaultTimeDisplayFormat(language),
		DefaultTimeZoneMode:       DefaultTimeZoneModeFloating,
		TodayPrimaryLookaheadDays: 3,
		DeadlineAwarenessDays:     14,
		WeatherCity:               "Tokyo",
	}
	return settings
}

func DefaultLocale(language LanguageCode) string {
	switch language {
	case LanguageSimplifiedChinese:
		return "zh-CN"
	case LanguageJapanese:
		return "ja-JP"
	default:
		return "en-US"
	}
}

func DefaultWeekStartDay(language LanguageCode) WeekStartDay {
	switch language {
	case LanguageSimplifiedChinese, LanguageJapanese:
		return WeekStartMonday
	default:
		return WeekStartSunday
	}
}

func DefaultTimeDisplayFormat(language LanguageCode) TimeDisplayFormat {
	switch language {
	case LanguageSimplifiedChinese, LanguageJapanese:
		return TimeDisplay24Hour
	default:
		return TimeDisplay12Hour
	}
}

func ValidateSettings(settings Settings) *Error {
	err := validationError("settings are invalid")

	if !IsValidLanguageCode(settings.Language) {
		err.addField(FieldLanguage, "invalid")
	}
	if settings.Locale == "" {
		err.addField(FieldLocale, "required")
	}
	if !IsValidWeekStartDay(settings.WeekStartDay) {
		err.addField(FieldWeekStartDay, "invalid")
	}
	if _, loadErr := time.LoadLocation(settings.TimeZone); loadErr != nil {
		err.addField(FieldTimeZone, "invalid")
	}
	if settings.DateDisplayFormat == "" {
		err.addField(FieldDateDisplayFormat, "required")
	}
	if !IsValidTimeDisplayFormat(settings.TimeDisplayFormat) {
		err.addField(FieldTimeDisplayFormat, "invalid")
	}
	if !IsValidDefaultTimeZoneMode(settings.DefaultTimeZoneMode) {
		err.addField(FieldDefaultTimeZoneMode, "invalid")
	}
	if settings.TodayPrimaryLookaheadDays < 0 {
		err.addField(FieldTodayPrimaryLookaheadDays, "must_be_non_negative")
	}
	if settings.DeadlineAwarenessDays < 0 {
		err.addField(FieldDeadlineAwarenessDays, "must_be_non_negative")
	}
	if settings.WeatherCity != "" && len(settings.WeatherCity) > 120 {
		err.addField(FieldWeatherCity, "too_long")
	}

	if err.hasFields() {
		return err
	}
	return nil
}
