import React, { useState } from 'react';
import { CalendarEvent, Event } from '../models';
import { Calendar, momentLocalizer, Views } from 'react-big-calendar';
import moment from 'moment';
import { convertEvents } from '../util';
import 'react-big-calendar/lib/css/react-big-calendar.css';
import './calendar-dark.css'
import EventView from './event';

interface CalendarViewProps {
  events: CalendarEvent[];
}

const CalendarView: React.FC<CalendarViewProps> = ({ events }) => {
  const localizer = momentLocalizer(moment);
  const calendarEvents = convertEvents(events);
  const [activeEvent, setActiveEvent] = useState<CalendarEvent| null>(null);

  return (
    <div style={{ display: 'flex', width: '100%' }}>
    <div style={{ flexGrow: 1 }}>
      <Calendar
        localizer={localizer}
        events={calendarEvents}
        style={{ height: "95vh" }}
        startAccessor="start"
        endAccessor="end"
        defaultView={Views.WEEK}
        className="dark-mode"
        onSelectEvent={(e: Event) => setActiveEvent(e.original)}
      />
    </div>
    {activeEvent && (
      <div style={{ width: '20%', padding: '10px' }}>
        <EventView event={activeEvent} showDescription={true} />
      </div>
    )}
  </div>
  );
};

export default CalendarView;