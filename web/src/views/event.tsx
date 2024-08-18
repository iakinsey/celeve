import React from 'react';
import { CalendarEvent } from '../models';
import ReactMarkdown from 'react-markdown';

interface EventViewProps {
  event: CalendarEvent;
  showDescription?: boolean;
}

const timeOptions: Intl.DateTimeFormatOptions = {
  month: 'long',
  day: 'numeric',
  year: 'numeric',
  hour: 'numeric',
  minute: 'numeric',
  hour12: true,
};

const Description: React.FC<{ text: string }> = ({ text }) => {
  return (
    <ReactMarkdown>{text}</ReactMarkdown>
  );
};

const EventView: React.FC<EventViewProps> = ({ event, showDescription }) => {
  const formatDate = (isoDate: string): string => {
    const date = new Date(isoDate);
    return date.toLocaleString('en-US', timeOptions);
  };

  return (
    <div>
        <div key={event.ID} style={styles.eventBlock}>
          <a href={event.OriginURL} style={styles.title} target="_blank" rel="noopener noreferrer">
            {event.Name}
          </a>
          <div style={styles.details}>
            <div style={styles.date}>{formatDate(event.StartTime)}</div>
            <div style={styles.tags}>
              {event.Tags.map((tag: string, index: number) => (
                <span key={index} style={styles.tag}>
                  {tag}
                </span>
              ))}
            </div>
            {
                showDescription &&
                <div style={styles.description}>
                    <Description text={event.Description} />
                </div>
            }
          </div>
        </div>
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
  description: {
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

export default EventView;
