import tamir from "../assets/tamir.jpg";
import { useState } from "react";
import "../../styles/about.scss"; // Import the SCSS file

export default function About() {
  const [contentID, setContentID] = useState(0);

  const content = [
    {
      title: "What is ",
      description: `Gobank is a secured money transfer app that allows you to create account and see your transactions history`,
      link: true,
      subtitle: "Gobank",
      contact: null,
    },
    {
      title: "Tamir Liebermann",
      description: `Passionate individual that loves creating, learning and designing. I am naturally curious, quietly confident and constantly improving myself each day. If I am not programming, I am either hanging out with my adorable dogs 🐶 or playing video games 🎮!`,
      link: null,
      subtitle: "Fullstack developer",
      contact: [
        "http://nguyntony.com/",
        "https://www.linkedin.com/in/nguyntony/",
        "https://github.com/nguyntony",
      ],
    },
   
  ];

  const clickTamir = () => {
    setContentID(1);
  };

  

  const defaultInfo = () => {
    setContentID(0);
  };

  return (
    <section className="aboutContainer">
      <div className="profileContainer">
        <div className="headshot one">
          <img src={tamir} alt="tamir" onClick={clickTamir} />
        </div>
        <div className="three" onClick={defaultInfo}>
          <img
            src="https://img.icons8.com/plasticine/100/000000/bank-card-back-side.png"
            alt="credit card"
          />
        </div>
      </div>

      <div className="aboutContentContainer">
        <div className="aboutCard">
          <div className="aboutContent">
            <div className="titleHeading">
              <h1>
                {content[contentID].title}
                {content[contentID].link && (
                  <span>
                    <a
                      href="https://github.com/tamir-liebermann/gobank"
                      target="_blank"
                      rel="noreferrer"
                    >
                      gobank
                    </a>
                    ?
                  </span>
                )}
              </h1>
              {content[contentID].subtitle && (
                <h3>{content[contentID].subtitle}</h3>
              )}
            </div>

            <div className="profileDescription">
              <p>
                {content[contentID].description}
                {content[contentID].link && (
                  <a
                    href="https://twintracker.me/"
                    target="_blank"
                    rel="noreferrer"
                  >
                    personal finance.
                  </a>
                )}
              </p>

              {content[contentID].contact && (
                <nav id="profileNav">
                  <ul>
                    <li>
                      <a
                        href={content[contentID].contact[0]}
                        target="_blank"
                        rel="noreferrer"
                      >
                        <i className="fas fa-address-card"></i>
                      </a>
                    </li>
                    <li>
                      <a
                        href={content[contentID].contact[1]}
                        target="_blank"
                        rel="noreferrer"
                      >
                        <i className="fab fa-linkedin"></i>
                      </a>
                    </li>
                    <li>
                      <a
                        href={content[contentID].contact[2]}
                        target="_blank"
                        rel="noreferrer"
                      >
                        <i className="fab fa-github"></i>
                      </a>
                    </li>
                  </ul>
                </nav>
              )}
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}
