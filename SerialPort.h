#pragma once

#include <boost/asio.hpp>
#include <boost/thread/thread.hpp>
#include <string>
#ifdef _WIN32
#include <windows.h>
#include <winsock2.h>
#else
#include <sys/ioctl.h>
#endif

using namespace boost;
class Serial {
private:
  boost::asio::io_service io;
  boost::asio::serial_port serial;
  boost::thread runner;
  boost::asio::streambuf buffer;
  bool quitFlag;
  std::function<void(std::string t)> cb;

  void set_dtr(asio::serial_port &serial, bool enable) {
#ifdef _WIN32
    DCB dcb;
    memset(&dcb, 0, sizeof(DCB));
    dcb.DCBlength = sizeof(DCB);
    GetCommState(serial.native_handle(), &dcb);
    dcb.fDtrControl = (enable) ? DTR_CONTROL_ENABLE : DTR_CONTROL_DISABLE;
    SetCommState(serial.native_handle(), &dcb);
#else
    if (enable) {
      ioctl(serial.native_handle(), TIOCMBIS, TIOCM_DTR)
    } else {
      ioctl(serial.native_handle(), TIOCMBIC, TIOCM_DTR)
    }
#endif
  }

public:
  Serial(std::function<void(std::string s)> cb)
      : serial(io), quitFlag(false), cb(cb){};

  ~Serial() {
    // Stop the I/O services
    io.stop();
    // Wait for the thread to finish
    runner.join();
  }

  bool connect(const std::string &port_name, int baud_rate = 9600) {
    using namespace boost::asio;
    serial.open(port_name);
    // Setup port
    serial.set_option(asio::serial_port_base::baud_rate(baud_rate));
    
    serial.set_option(asio::serial_port_base::stop_bits(
      asio::serial_port_base::stop_bits::one));

    serial.set_option(
      asio::serial_port_base::parity(asio::serial_port_base::parity::none));
    serial.set_option(asio::serial_port_base::character_size(8));

    set_dtr(serial, true);

#ifdef _WIN32
    PurgeComm(serial.native_handle(), PURGE_RXCLEAR);
#else
    tcflush(serial.native_handle(), TCIFLUSH);
#endif

    if (serial.is_open()) {
      // Start io-service in a background thread.
      // boost::bind binds the ioservice instance
      // with the method call
      runner = boost::thread(boost::bind(&boost::asio::io_service::run, &io));

      startReceive();
    }

    return serial.is_open();
  }

  void startReceive() {
    using namespace boost::asio;
    // Issue a async receive and give it a callback
    // onData that should be called when "\r\n"
    // is matched.
    async_read_until(serial, buffer, "\n",
                     boost::bind(&Serial::onData, this, _1, _2));
  }

  void send(const std::string &text) {
    boost::asio::write(serial, boost::asio::buffer(text));
  }

  void onData(const boost::system::error_code &e, std::size_t size) {
    if (!e) {
      std::istream is(&buffer);
      std::string data(size, '\0');
      is.read(&data[0], size);

      cb(data);
      // If we receive quit()\r\n indicate
      // end of operations
      // quitFlag = (data.compare("quit()\r\n") == 0);
    };

    startReceive();
  };

  bool quit() {
    return quitFlag;
  }
};