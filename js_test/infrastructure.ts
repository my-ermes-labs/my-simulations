export const infrastructure = {
  "areaIdentifiers": ["states", "cities"],
  "areas": [
    {
      /***************************** AREA 1 ***********************************/
      // Hosted aws ec2 instance in Milan, Italy.
      "areaName": "Italy",
      "host": "<ip-1>",
      // We can't need know the specific location of the host, so we'll approximate 
      // it to the one of the centre of Milan.
      "geoCoordinates": { "latitude": 45.4642, "longitude": 9.1900 },
      "tags": { "tag": "ec2-instance" },
      // We expect this machine to have triple the compute power of the reference machine (my laptop).
      "resources": { "reference_compute_unit": 3 },
      // The sub-areas.
      "areas": [
        {
          /**************************** AREA 1.1 ******************************/
          "areaName": "Ravenna (PC)",
          // Since 2 machines are hosted with the same IP, we'll differentiate them by port.
          "host": "<ip-1.1>",
          // Geo coordinates of Ravenna, Italy.
          "geoCoordinates": { "latitude": 44.4173, "longitude": 12.2019 },
          "tags": { "tag": "Personal PC" },
          // We expect this machine to have double the compute power of the reference machine (my laptop).
          "resources": { "reference_compute_unit": 2 },
        },
        {
          /**************************** AREA 1.2 ******************************/
          "areaName": "Ravenna (Laptop)",
          // Since 2 machines are hosted with the same IP, we'll differentiate them by port.
          "host": "<ip-1.2>",
          // Geo coordinates of Ravenna, Italy.
          "geoCoordinates": { "latitude": 44.4173, "longitude": 12.2019 },
          "tags": { "tag": "Personal Laptop" },
          // We expect this machine to have double the compute power of the reference machine (my laptop).
          "resources": { "reference_compute_unit": 1 },
        },
      ]
    },
  ]
} as const;