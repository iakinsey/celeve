import React from 'react';
import { CalendarEvent } from '../models';
import EventView from './event';

interface ListingViewProps {
  events: CalendarEvent[];
}

const ListingView: React.FC<ListingViewProps> = ({ events }) => {
  return (
    <div>
      {events.map((event) => (
        <EventView event={event} />
      ))}
    </div>
  );
};

const styles = {
  eventBlock: {
    border: '1px solid #0e0e0e',
    borderRadius: '5px',
    padding: '10px',
    marginBottom: '10px',
    backgroundColor: '#2d2d2d',
    cursor: 'pointer',
  },
  title: {
    fontSize: '18px',
    fontWeight: 'bold',
    textDecoration: 'none',
    color: '#b8b8b8',
  },
  details: {
    marginTop: '5px',
  },
  date: {
    fontSize: '14px',
    color: '#666',
  },
  tags: {
    marginTop: '5px',
  },
  tag: {
    display: 'inline-block',
    backgroundColor: '#848484',
    borderRadius: '3px',
    padding: '2px 6px',
    marginRight: '5px',
    fontSize: '12px',
    color: '#070707',
  },
};

export default ListingView;
