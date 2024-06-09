import React from "react";
import { BrowserRouter as Router, Route, Routes } from "react-router-dom";
import Nav from "./Nav";
import CreateAccount from "./components/CreateAccount";
import Login from "./components/Login";
import TransactionsHistory from "./components/TransactionsHistory";
import About from "./components/About";
import Transfer from "./components/Transfer";

const App = () => {
  return (
    <Router>
      <div>
        <Nav />
        <Routes>
          <Route path="/create" component={CreateAccount} />
          <Route path="/login" component={Login} />
          <Route
            path="/account/transactions/:id"
            component={TransactionsHistory}
          />
          <Route path="/account/transfer/:id" component={Transfer} />
          <Route path="/about" component={About} />
          <Route path="/" component={() => <div>Home</div>} />
        </Routes>
      </div>
    </Router>
  );
};

export default App;
