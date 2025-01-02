export type PageData = LocalData & ExternalData;

export type ExternalData = Weather;

export type LocalData = BoilerData & SensorData & SensorRangeData & StateHistory;

export type Weather = {
	outsideTemperatureSeries: Measure[];
};

export type BoilerData = {
	boiler: Boiler;
};

export type StateHistory = {
	switchHistory: SwitchSample[];
	overheatingProtectionHistory: OverheatingProtectionSample[];
}

export type SensorData = {
	currentTemperature?: Measure;
};

export type SensorRangeData = {
	temperatureSeries: Measure[];
	humiditySeries: Measure[];
};

export type Boiler = {
	state: State;
	minTemp: number;
	maxTemp: number;
	rules: Rule[];
};

export type Rule = RuleInput & {
	id: string;
	stoppedTime: Date;
	isActive: boolean;
};

export type RuleInput = {
	start: string;
	duration: string;
	delay: string;
	targetTemp: number;
	repeatDays: number[];
};

export type State = 'OFF' | 'ON';

export type Measure = {
	time: string;
	value: number;
};

export type SwitchSample = {
	time: string;
	state: string;
}

export type OverheatingProtectionSample = {
	time: string;
	isActive: boolean;
}

export type ColorBand = {
	from: string;
	to: string;
	color: string;
}
