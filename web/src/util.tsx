import { CalendarEvent, Event } from "./models";

export async function request(url: string, body?: any) {
    url = "http://localhost:8989" + url
    const headers = {
        'Content-Type': 'application/json',
    };

    const options: RequestInit = {
        method: 'POST',
        headers,
        body: body ? JSON.stringify(body) : undefined,
    };

    try {
        const response = await fetch(url, options);
        if (!response.ok) {
            throw new Error(`HTTP error! Status: ${response.status}`);
        }
        const data = await response.json();
        return data;
    } catch (error) {
        console.error('Error making HTTP request:', error);
        throw error;
    }
}

export function convertEvents(events: CalendarEvent[]): Event[] {
    return events.map((event) => ({
        id: event.ID,
        title: event.Name,
        start: new Date(event.StartTime),
        end: new Date(event.EndTime),
        original: event,
    }));
}