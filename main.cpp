#include <boost/asio.hpp>
#include <boost/asio/serial_port.hpp>
#include <boost/thread/thread.hpp>
#include <fstream>
#include <iostream>
#include <string>

#include "serialport.h"

using namespace boost;
int main() {
  std::ofstream myfile;
  myfile.open("log.txt");
  Serial serial([&](std::string t) {
    std::cout << t << std::endl;
    myfile << t;

  });
  try {
    if (serial.connect("COM12", 115200)) {
      std::cout << "Port is open." << std::endl;
    } else {
      std::cout << "Port open failed." << std::endl;
    }
  } catch (boost::system::system_error &error) {
    std::cerr << "PORT NOT AVAILABLE" << std::endl;
    return -1;
  }

  while (!serial.quit()) {
    // Do something...
    std::cout << "IK DOE LEKKER OOK DINGEN HOOR\n";

    boost::this_thread::sleep_for(boost::chrono::milliseconds(5000));
  }
  myfile.close();

  return 0;
}
