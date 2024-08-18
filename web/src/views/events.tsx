import React, { Component } from 'react';
import { request } from '../util';
import { CalendarEvent } from '../models';
import ListingView from './listing';
import CalendarView from './calendar';
import Select, { StylesConfig } from 'react-select';

interface MainViewProps {
    start: number;
    end: number;
    limit: number;
    offset: number;
}

interface MainViewState {
    start: number;
    end: number;
    limit: number;
    offset: number;
    events: CalendarEvent[];
    tags: string[];
    tagChoices: TagChoice[];
    view: string;
}

interface TagChoice {
    value: string;
    label: string;
}

export default class MainView extends Component<MainViewProps, MainViewState> {
    constructor(props: MainViewProps) {
        super(props);
        
        this.state = {
            start: props.start,
            end: props.end,
            limit: props.limit,
            offset: props.offset,
            events: [],
            tags: [],
            tagChoices: [],
            view: 'calendar',
        };
    }

    setView(val: string) {
        this.setState({view: val})
    }

    componentDidMount() {
        this.getEvents()
        this.getTags()
    }

    componentDidUpdate() {
        this.getEvents()
    }

    async getEvents() {
        let events = await request("/events", {
            start: this.state.start,
            end: this.state.end,
            limit: this.state.limit,
            offset: this.state.offset,
            tags: this.state.tags
        })

        this.setState({
            events: events || []
        })
    }
    
    async getTags() {
        let tags: string[] = await request("/tags") || []

        this.setState({
            tagChoices: tags.map((tag: string) => ({
                value: tag,
                label: tag
              }))
              
        })
    }

    onChange(choices: TagChoice[]) {
        this.setState({
            tags: choices.map((choice: TagChoice) => choice.value)
        })
    }

    render() {
        return (
            <div>
                <div>
                    <Select styles={customStyles} isMulti={true} options={this.state.tagChoices} onChange={(a: any) => this.onChange(a)} />
                </div>
                {this.state.view === 'listing' ? (
                    <ListingView events={this.state.events} />
                ) : (
                    <CalendarView events={this.state.events} />
                )}
          </div>
        );
    }
}

const customStyles = {
    control: (provided: any) => ({
      ...provided,
      backgroundColor: '#333',
      border: 'none',
      boxShadow: 'none',
    }),
    menu: (provided: any) => ({
      ...provided,
      backgroundColor: '#333',
      color: '#fff',
      zIndex: 1000,
      marginTop: '0', // Removes the gap
      position: 'absolute', // Ensures the menu is directly attached to the control
    }),
    menuPortal: (base: any) => ({ 
      ...base, 
      zIndex: 1000,
      top: '100%', // Ensures the menu appears directly below the control
    }),
    option: (provided: any, state: any) => ({
      ...provided,
      backgroundColor: state.isSelected ? '#555' : '#333',
      color: state.isSelected ? '#fff' : '#ddd',
      '&:hover': {
        backgroundColor: '#444',
      },
    }),
    multiValue: (provided: any) => ({
      ...provided,
      backgroundColor: '#555',
      color: '#fff',
    }),
    multiValueLabel: (provided: any) => ({
      ...provided,
      color: '#fff',
    }),
    multiValueRemove: (provided: any) => ({
      ...provided,
      color: '#fff',
      '&:hover': {
        backgroundColor: '#777',
        color: '#fff',
      },
    }),
    singleValue: (provided: any) => ({
      ...provided,
      color: '#fff',
    }),
    placeholder: (provided: any) => ({
      ...provided,
      color: '#888',
    }),
  };
/*
const customStyles = {
    control: (provided: any) => ({
      ...provided,
      backgroundColor: '#333',
      border: 'none',
      boxShadow: 'none',
    }),
    menu: (provided: any) => ({
      ...provided,
      backgroundColor: '#333',
      color: '#fff',
      zIndex: 1000,
    }),
    menuPortal: (base: any) => ({ 
      ...base, 
      zIndex: 1000,
    }),
    option: (provided: any, state: any) => ({
      ...provided,
      backgroundColor: state.isSelected ? '#555' : '#333',
      color: state.isSelected ? '#fff' : '#ddd',
      '&:hover': {
        backgroundColor: '#444',
      },
    }),
    multiValue: (provided: any) => ({
      ...provided,
      backgroundColor: '#555',
      color: '#fff',
    }),
    multiValueLabel: (provided: any) => ({
      ...provided,
      color: '#fff',
    }),
    multiValueRemove: (provided: any) => ({
      ...provided,
      color: '#fff',
      '&:hover': {
        backgroundColor: '#777',
        color: '#fff',
      },
    }),
    singleValue: (provided: any) => ({
      ...provided,
      color: '#fff',
    }),
    placeholder: (provided: any) => ({
      ...provided,
      color: '#888',
    }),
  };
  */