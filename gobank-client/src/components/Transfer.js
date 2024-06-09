import React, { useState } from "react";
import { useParams } from "react-router-dom";
import axios from "axios";

const Transfer = () => {
  const { id } = useParams();
  const [formData, setFormData] = useState({
    amount: "",
    recipientId: "",
  });

  const handleChange = (e) => {
    setFormData({ ...formData, [e.target.name]: e.target.value });
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    try {
      const response = await axios.post(
        `http://localhost:8080/account/transfer/${id}`,
        formData
      );
      console.log("Transfer successful:", response.data);
    } catch (error) {
      console.error("Error making transfer:", error);
    }
  };

  return (
    <form onSubmit={handleSubmit}>
      <input
        type="text"
        name="amount"
        placeholder="Amount"
        onChange={handleChange}
      />
      <input
        type="text"
        name="recipientId"
        placeholder="Recipient ID"
        onChange={handleChange}
      />
      <button type="submit">Transfer</button>
    </form>
  );
};

export default Transfer;
