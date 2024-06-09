import React from "react";
import { Link, useLocation } from "react-router-dom";

const Nav = () => {
  const location = useLocation();
  const currentPath = location.pathname;

  const underline = {
    textDecoration: "underline",
    textDecorationColor: "#ff778f",
    textDecorationThickness: "0.3rem",
  };

  const loginUnderline = currentPath === "/login" ? underline : null;
  const signupUnderline = currentPath === "/create" ? underline : null;
  const aboutUnderline = currentPath === "/About" ? underline : null;

  return (
    <nav className="homeNav">
      <h2>GoBank</h2>
      <ul className="nav-list">
        <Link to="/about">
          <li style={aboutUnderline}>About</li>
        </Link>
        <Link to="/login">
          <li style={loginUnderline}>Login</li>
        </Link>
        <Link to="/create">
          <li style={signupUnderline}>Create account</li>
        </Link>
      </ul>
    </nav>
  );
};

export default Nav;
