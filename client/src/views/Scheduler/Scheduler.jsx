import React, { useState, useEffect } from 'react';

// @material-ui/core
import { makeStyles, withStyles } from '@material-ui/core/styles';
import Paper from '@material-ui/core/Paper';
import Select from '@material-ui/core/Select';
import MenuItem from '@material-ui/core/MenuItem';
// import InputLabel from '@material-ui/core/InputLabel';
import OutlinedInput from '@material-ui/core/OutlinedInput';

import TimeTableCell from '../../components/Scheduler/TimeTableCell';
import DayScaleCell from '../../components/Scheduler/DayScaleCell';
import AppointmentContent from '../../components/Scheduler/AppointmentContent';

// core components
import GridContainer from '../../components/Grid/GridContainer';
import GridItem from '../../components/Grid/GridItem';

import { ViewState, EditingState } from '@devexpress/dx-react-scheduler';
import {
  Scheduler,
  Toolbar,
  Resources,
  Appointments,
  WeekView,
  DateNavigator,
  DayView,
  ViewSwitcher,
  TodayButton,
  AppointmentTooltip,
  EditRecurrenceMenu,
  DragDropProvider,
  AppointmentForm,
  CurrentTimeIndicator,
  ConfirmationDialog,
} from '@devexpress/dx-react-scheduler-material-ui';
import { connectProps } from '@devexpress/dx-react-core';

// const useStyles = makeStyles(schedulerStyles);

const useStyles = makeStyles(theme => ({
  noseeum: {
    display: 'none',
    visibility: 'hidden',
  },
  root: {
    width: 120,
  },
}));

const messages = {
  weeksOnLabel: 'week(s).',
  moreInformationLabel: '',
};

const today = new Date();

const DateEditor = ({ excludeTime, ...restProps }) => {
  const dateFormat = excludeTime ? 'YYYY-MM-DD' : 'YYYY-MM-DD hh:mm A';

  return (
    <AppointmentForm.DateEditor
      {...restProps}
      excludeTime={excludeTime}
      format={dateFormat}
    />
  );
};

const TextEditor = props => {
  // eslint-disable-next-line react/destructuring-assignment
  if (props.type === 'multilineTextEditor') {
    return null;
  }
  return <AppointmentForm.TextEditor {...props} />;
};

const BooleanEditor = props => {
  if (props.label !== 'Repeat') {
    return null;
  }
  return <AppointmentForm.BooleanEditor {...props} />;
};

const WeeklyRecurrence = ({ ...restProps }) => {
  const classes = useStyles();
  return (
    <AppointmentForm.WeeklyRecurrenceSelector
      {...restProps}
      readOnly={true}
      className={classes.noseeum}
    />
  );
};

const AgentGroupSelector = ({ ...restProps }) => {
  const { data, originalApptList, setData, groupView, setGroupView } =
    restProps;

  const handleChange = e => {
    setGroupView(e.target.value);
  };

  const classes = useStyles();
  let options = [
    { value: 'everyone', label: 'Everyone' },
    { value: 'Support', label: 'Support Specialists' },
    { value: 'Support Engineers', label: 'Support Engineers' },
  ];

  return (
    <Toolbar.FlexibleSpace style={{ margin: '0 25px 0 auto' }}>
      <Select
        style={{ width: '200px', fontSize: '14px', textTransform: 'uppercase' }}
        value={groupView}
        className={classes.root}
        input={<OutlinedInput margin='dense' />}
        onChange={handleChange}
      >
        {options.map((option, index) => (
          <MenuItem key={index} value={option.value}>
            {option.label}
          </MenuItem>
        ))}
      </Select>
    </Toolbar.FlexibleSpace>
  );
};

const SchedulerContainer = () => {
  const [originalApptList, setOriginalApptList] = useState([]);
  const [data, setData] = useState([]);
  const [currentDate, setCurrentDate] = useState(today);
  const [currentViewName, setCurrentViewName] = useState('Week');
  const [agentResources, setAgentResources] = useState([]);
  const [groupView, setGroupView] = useState('everyone');

  const mainResourceName = 'agent';

  useEffect(() => {
    setData([
      ...originalApptList.filter(item => {
        if (item.group === groupView) {
          return item;
        }
        if (groupView === 'everyone') {
          return item;
        }
      }),
    ]);
  }, [groupView]);

  const FlexibleSpace = connectProps(AgentGroupSelector, () => {
    return {
      originalApptList,
      data,
      groupView,
      setData,
      setGroupView,
    };
  });

  useEffect(() => {
    let options = {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
      },
      credentials: 'include',
    };

    fetch(`/api/v1/appointments`, options)
      .then(response => response.json())
      .then(({ appointments }) => {
        if (appointments == null) {
          setOriginalApptList([]);
          setData([]);
          return;
        }

        // convert appointment key to an id the scheduler can use
        // convert date strings to dates the scheduler can use
        let tempAppts = [];
        appointments.map(appt => {
          appt.id = appt.key;
          appt.startDate = new Date(appt.startDate);
          appt.endDate = new Date(appt.endDate);
          tempAppts.push(appt);
        });

        setOriginalApptList(tempAppts);
        setData(tempAppts);
        setGroupView('everyone');
      })
      .catch(error => console.log(error));

    fetch(`/api/v1/agent/active`, options)
      .then(response => response.json())
      .then(({ agents }) => {
        if (agents === null) {
          console.warn('no active agents');
          return;
        }
        // Resoureces must have an `id`. The server provides a `key.`
        // create a temp list, map over the original list and add an `id`
        // push to temp list; and then use temp list as the resources.
        let tempAgentList = [];
        agents.map(agent => {
          agent.id = agent.key;
          tempAgentList.push(agent);
        });
        setAgentResources([
          {
            fieldName: 'agent',
            title: 'Agent',
            allowMultiple: false,
            instances: tempAgentList,
          },
        ]);
      })
      .catch(error => console.log(error));
  }, []);

  const callAppointmentsRoute = (obj, action) => {
    let options = {
      method: '',
      headers: {
        'Content-Type': 'application/json',
      },
    };

    if (action === 'delete') {
      options.method = 'DELETE';
      fetch(`/api/v1/appointments/${obj}`, options) // TODO: need to check this `obj`  value works with the new endpoint
        .then(response => response.json())
        .then(({ message }) => {
          if (message == 'deleted') {
            let newData = data.filter(appointment => appointment.key !== obj);
            setData(newData);
            setOriginalApptList(newData);
            return;
          }
          let error = new Error(message);
          throw error;
        })
        .catch(error => console.log(error));
    }

    if (action === 'add') {
      console.log('adding');
      options.method = 'POST';
      options.body = JSON.stringify(obj);
      console.log('payload: ', options.body);
      fetch(`/api/v1/appointments`, options)
        .then(response => response.json())
        .then(({ appointment }) => {
          appointment.id = appointment.key;
          setData([...data, appointment]);
          setOriginalApptList([...originalApptList, appointment]);
        })
        .catch(error => console.log(error));
    }

    if (action === 'change') {
      console.log('changing');
      let objKey = Object.getOwnPropertyNames(obj).toString();
      let payload = data.find(x => x.key === objKey);
      let keys = Object.keys(obj[objKey]);

      keys.map(key => {
        payload[key] = obj[objKey][key];
      });

      options.method = 'PATCH';
      options.body = JSON.stringify(payload);
      console.log('payload: ', options.body);
      fetch(`/api/v1/appointments/${objKey}`, options)
        .then(response => response.json())
        .then(({ appointment }) => {
          let newData = data.map(appt => {
            if (appt.key === appointment.key) {
              return { ...appt, ...appointment };
            }
            return appt;
          });
          setData(newData);
          setOriginalApptList(newData);
        })
        .catch(error => console.log(error));
    }
  };

  const commitChanges = ({ added, changed, deleted }) => {
    if (deleted) {
      callAppointmentsRoute(deleted, 'delete');
    }

    if (added) {
      console.log('added: ', added);
      callAppointmentsRoute(added, 'add');
    }

    if (changed) {
      console.log('changed: ', changed);
      callAppointmentsRoute(changed, 'change');
    }
  };

  const onCurrentDateChange = e => {
    setCurrentDate(e);
  };

  const onCurrentViewNameChange = e => {
    setCurrentViewName(e);
  };

  return (
    <GridContainer>
      <GridItem>
        <Paper style={{ height: '90vh', marginBottom: '10px' }} elevation={3}>
          <Scheduler data={data} locale={'en-US'}>
            <ViewState
              currentDate={currentDate}
              onCurrentDateChange={onCurrentDateChange}
              currentViewName={currentViewName}
              onCurrentViewNameChange={onCurrentViewNameChange}
            />
            <EditingState onCommitChanges={commitChanges} />
            <EditRecurrenceMenu />
            <DayView startDayHour={0} endDayHour={24} />
            <WeekView
              startDayHour={0}
              endDayHour={24}
              timeTableCellComponent={TimeTableCell}
              dayScaleCellComponent={DayScaleCell}
            />
            <Toolbar flexibleSpaceComponent={FlexibleSpace} />
            <DateNavigator />
            <ViewSwitcher />
            <TodayButton />
            <Appointments appointmentContentComponent={AppointmentContent} />
            <Resources
              data={agentResources}
              mainResourceName={mainResourceName}
            />
            <ConfirmationDialog />
            <AppointmentTooltip
              showOpenButton
              showDeleteButton
              showCloseButton
            />
            <AppointmentForm
              dateEditorComponent={DateEditor}
              textEditorComponent={TextEditor}
              booleanEditorComponent={BooleanEditor}
              weeklyRecurrenceSelectorComponent={WeeklyRecurrence}
              messages={messages}
            />
            <DragDropProvider />
            <CurrentTimeIndicator
              updateInterval={300000}
              shadePreviousAppointments={true}
              shadePreviousCells={true}
            />
          </Scheduler>
        </Paper>
      </GridItem>
    </GridContainer>
  );
};

export default SchedulerContainer;
