export type PageData = BoilerData & SensorData & SensorRangeData;

export type BoilerData = {
	boiler: Boiler;
};

export type SensorData = {
	sensor?: Measure;
};

export type SensorRangeData = {
	sensorRange: Measure[];
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
	targetTemp: number;
	repeatDays: number[];
};

export type State = 'OFF' | 'ON';

export type Measure = {
	timestamp: string;
	value: number;
};