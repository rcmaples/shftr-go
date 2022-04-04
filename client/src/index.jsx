import React from 'react';
import ReactDOM from 'react-dom';
import { Provider } from 'react-redux';
import { BrowserRouter, Route, Switch, Redirect } from 'react-router-dom';

// import * as FullStory from '@fullstory/browser';
// FullStory.init({ orgId: '164EXE' });

// core components
import Admin from './layouts/Admin';
import './styles/css/main.css';

ReactDOM.render(
  <BrowserRouter>
    <Switch>
      <Route path='/admin' component={Admin} />
      <Redirect from='/' to='/admin' />
    </Switch>
  </BrowserRouter>,
  document.getElementById('root')
);
