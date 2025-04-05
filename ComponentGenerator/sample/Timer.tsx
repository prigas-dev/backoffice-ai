import { useEffect, useState } from "react";

export function Component() {
  const [formattedTime, setFormattedTime] = useState("...");

  useEffect(() => {
    const intervalId = setInterval(() => {
      const formattedTime = formatDateIntl(new Date());
      setFormattedTime(formattedTime);
    }, 50);

    return () => {
      clearInterval(intervalId);
    };
  }, []);
  return <div className="container">Now: {formattedTime}</div>;
}

function formatDateIntl(date: Date) {
  const datePart = new Intl.DateTimeFormat("en-GB", {
    day: "2-digit",
    month: "2-digit",
    year: "numeric",
  }).format(date);

  const timePart = new Intl.DateTimeFormat("en-GB", {
    hour: "2-digit",
    minute: "2-digit",
    second: "2-digit",
    hour12: false, // set to true for 12-hour format
  }).format(date);

  return `${datePart} ${timePart}`;
}
