import React, { useState, useEffect, useRef } from 'react';
import * as FullStory from '@fullstory/browser';
import { makeStyles } from '@material-ui/core/styles';
import Table from '@material-ui/core/Table';
import TableBody from '@material-ui/core/TableBody';
import TableCell from '@material-ui/core/TableCell';
import TableContainer from '@material-ui/core/TableContainer';
import TableRow from '@material-ui/core/TableRow';
import Checkbox from '@material-ui/core/Checkbox';
import Button from '@material-ui/core/Button';
import SyncIcon from '@material-ui/icons/Sync';
import Switch from '@material-ui/core/Switch';

import TableToolbar from '../../components/TableToolbar/TableToolbar';
import AgentTableHeader from './AgentTableHeader';
import AgentTableFooter from '../../components/AgentTableFooter/AgentTableFooter';
import { Event, InsertComment } from '@material-ui/icons';

const useStyles = makeStyles(theme => ({
  root: {
    width: '100%',
  },
  paper: {
    width: '100%',
    marginBottom: theme.spacing(2),
  },
  table: {
    minWidth: 750,
  },
  visuallyHidden: {
    border: 0,
    clip: 'rect(0 0 0 0)',
    height: 1,
    margin: -1,
    overflow: 'hidden',
    padding: 0,
    position: 'absolute',
    top: 20,
    width: 1,
  },
}));

function descendingComparator(a, b, orderBy) {
  if (b[orderBy] < a[orderBy]) {
    return -1;
  }
  if (b[orderBy] > a[orderBy]) {
    return 1;
  }
  return 0;
}

function stableSort(array, comparator) {
  const stabilizedThis = array.map((el, index) => [el, index]);
  stabilizedThis.sort((a, b) => {
    const order = comparator(a[0], b[0]);
    if (order !== 0) return order;
    return a[1] - b[1];
  });
  return stabilizedThis.map(el => el[0]);
}

function getComparator(order, orderBy) {
  return order === 'desc'
    ? (a, b) => descendingComparator(a, b, orderBy)
    : (a, b) => -descendingComparator(a, b, orderBy);
}

function updateArray(arr, obj) {
  const index = arr.findIndex(el => el.key === obj.key);
  if (index === -1) {
    return [...arr, obj];
  } else {
    arr.splice(index, 1);
    return arr;
  }
}

const AgentTable = () => {
  const classes = useStyles();
  const [order, setOrder] = useState('desc');
  const [orderBy, setOrderBy] = useState('activated');
  const [selected, setSelected] = useState([]);
  const [agentList, setAgentList] = useState([]);
  const [updates, setUpdates] = useState([]);
  // const [currentUser, setCurrentUser] = useState({});

  useEffect(() => {
    // let cookieToken = findToken();
    // if (cookieToken) {
    //   let decodedUser = jwt_decode(cookieToken);
    //   let { key, name, email, org } = decodedUser;
    //   setCurrentUser({ key, name, email, org });
    // }

    let options = {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
      },
    };
    fetch(`/api/v1/agents`, options)
      .then(response => {
        const { status, statusText } = response;
        if (status >= 200 && status < 300) {
          return response.json();
        } else {
          throw { status, statusText };
        }
      })
      .then(({ agents }) => {
        if (agents == null) {
          console.log('no agents found');
          return;
        }
        setAgentList(agents);
      })
      .catch(error => console.warn(error));
  }, []);

  const handleRequestSort = (event, property) => {
    const isAsc = orderBy === property && order === 'asc';
    setOrder(isAsc ? 'desc' : 'asc');
    setOrderBy(property);
  };

  const handleSelectAllClick = event => {
    if (event.target.checked) {
      const newSelecteds = agentList.map(n => n);
      setSelected(newSelecteds);
      return;
    }
    setSelected([]);
  };

  const handleSelectSingleClick = (event, name) => {
    if (event.target.key === 'activeAgent') {
      return;
    }

    const selectedIndex = selected.indexOf(name);
    let newSelected = [];

    if (selectedIndex === -1) {
      newSelected = newSelected.concat(selected, name);
    } else if (selectedIndex === 0) {
      newSelected = newSelected.concat(selected.slice(1));
    } else if (selectedIndex === selected.length - 1) {
      newSelected = newSelected.concat(selected.slice(0, -1));
    } else if (selectedIndex > 0) {
      newSelected = newSelected.concat(
        selected.slice(0, selectedIndex),
        selected.slice(selectedIndex + 1)
      );
    }

    setSelected(newSelected);
  };

  const handlePauseAgent = (event, agent) => {
    event.preventDefault();
    let action = '';
    let paused = event.target.checked;
    let { key, name, defaultZendeskGroupName } = agent;

    if (paused == true) {
      action = 'Paused Agent';
    }

    if (paused == false) {
      action = 'Unpaused Agent';
    }

    // TODO: REIMPLEMENT FS.EVENT
    // FullStory.event(action, {
    //   currentUserName_str: currentUser.name,
    //   currentUserEmail_str: currentUser.email,
    //   modifiedAgentId_str: id,
    //   modifiedAgentName: name,
    //   modifiedAgentZendeskGroup_str: defaultZendeskGroupName,
    // });

    let theUpdate = {
      key,
      paused,
    };

    let options = {
      method: 'PATCH',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(theUpdate),
      credentials: 'include',
    };

    fetch(`/api/v1/agent/pause`, options)
      .then(response => response.json())
      .then(({ agent }) => {
        if (typeof agent.paused != undefined) {
          let pausedUpdates = agentList.map(item => {
            if (item.key == agent.key) {
              return { ...item, paused: agent.paused };
            }
            return item;
          });
          setAgentList(pausedUpdates);
        }
      })
      .catch(error => console.warn(error));
  };

  const handleActivateClick = (event, agent) => {
    event.preventDefault();
    const { key } = agent;

    let newAgent = {
      ...agent,
      activated: !agent.activated,
    };

    let activeAgents = agentList.map(item => {
      if (item.key == key) {
        return { ...item, activated: !item.activated };
      }
      return item;
    });

    let updatedAgents = updateArray(updates, newAgent);
    setUpdates(updatedAgents);
    setAgentList(activeAgents);
  };

  const handleSaveUpdates = event => {
    event.preventDefault();
    let options = {
      method: 'PATCH',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(updates),
      credentials: 'include',
    };

    fetch(`/api/v1/agent`, options)
      .then(response => response.json())
      .then(data => {
        setUpdates([]);
      })
      .catch(error => console.warn(error));
  };

  const handleSyncClick = event => {
    event.preventDefault();
    let options = {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
    };
    fetch(`/api/v1/agents`, options)
      .then(response => response.json())
      .then(({ agents }) => {
        setAgentList([...agentList, ...agents]);
      })
      .catch(error => console.log(error));
  };

  const handleDeleteAgents = event => {
    event.preventDefault();
    let options = {
      method: 'DELETE',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(selected),
      credentials: 'include',
    };

    fetch(`/api/v1/agent`, options)
      .then(response => response.json())
      .then(data => {
        let newList = agentList;
        selected.map(item => {
          const index = newList.findIndex(el => el.key === item.key);
          if (index !== -1) {
            newList.splice(index, 1);
          }
        });

        setAgentList(newList);

        setSelected([]);
      })
      .catch(error => console.warn(error));
  };

  const isSelected = agent => selected.indexOf(agent) !== -1;

  return (
    <div className={classes.root}>
      {selected.length > 0 ? (
        <TableToolbar
          numSelected={selected.length}
          action='deleteAgents'
          onChildClick={handleDeleteAgents}
        />
      ) : (
        ''
      )}

      <TableContainer
        style={{
          height: '60vh',
          marginBottom: '10px',
          overflowY: 'scroll',
        }}
      >
        <Table
          stickyHeader
          className={classes.table}
          aria-labelledby='tableTitle'
          size='medium'
          aria-label='enhanced table'
        >
          <AgentTableHeader
            classes={classes}
            numSelected={selected.length}
            order={order}
            orderBy={orderBy}
            onSelectAllClick={handleSelectAllClick}
            onRequestSort={handleRequestSort}
            rowCount={agentList.length}
          />
          <TableBody>
            {stableSort(agentList, getComparator(order, orderBy)).map(
              (agent, index) => {
                const isItemSelected = isSelected(agent);
                const labelId = `enhanced-table-checkbox-${index}`;
                const zendeskGroupName = agent.defaultZendeskGroupName;
                return (
                  <TableRow
                    hover
                    role='checkbox'
                    aria-checked={isItemSelected}
                    tabIndex={-1}
                    key={agent.name}
                    selected={isItemSelected}
                  >
                    <TableCell padding='checkbox'>
                      <Checkbox
                        onClick={event => handleSelectSingleClick(event, agent)}
                        checked={isItemSelected}
                        inputProps={{ 'aria-labelledby': labelId }}
                      />
                    </TableCell>
                    <TableCell
                      component='th'
                      id={labelId}
                      scope='row'
                      padding='none'
                    >
                      {agent.name}
                    </TableCell>
                    <TableCell id={zendeskGroupName} align='right'>
                      {zendeskGroupName}
                    </TableCell>
                    <TableCell align='right'>
                      <Checkbox
                        id='activeAgent'
                        checked={agent.activated}
                        value={agent.activated}
                        onClick={event => handleActivateClick(event, agent)}
                      />
                    </TableCell>
                    <TableCell align='center'>
                      <Switch
                        size='small'
                        checked={agent.paused}
                        onChange={event => handlePauseAgent(event, agent)}
                        disabled={!agent.activated}
                      />
                    </TableCell>
                  </TableRow>
                );
              }
            )}
          </TableBody>
        </Table>
      </TableContainer>
      {updates.length > 0 ? (
        <TableToolbar
          numSelected={updates.length}
          action='makeUpdates'
          onChildClick={handleSaveUpdates}
        />
      ) : (
        ''
      )}
      <Button onClick={handleSyncClick}>
        <SyncIcon /> Pull data from zendesk
      </Button>
    </div>
  );
};

export default AgentTable;
