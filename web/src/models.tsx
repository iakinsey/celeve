export type CalendarEvent = {
	ID:          string;
	Name:        string;
	StartTime:   string;
	EndTime:     string;
	Location:    string;
	Description: string;
	OriginURL:   string;
	Tags:        string[];
	Processed:   boolean;
	Relevant:    boolean;
};

export type Event = {
	id: string;
	title: string;
	start: Date;
	end: Date;
	original: CalendarEvent
}