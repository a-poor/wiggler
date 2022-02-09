import React, { useEffect, useState } from 'react';
import './App.css';
import 'antd/dist/antd.css';

import { 
  Typography,
  Button,
  Spin,
  Divider,
  Space,
  Form,
  Slider,
  Tooltip,
  message,
} from 'antd';

import {
  DownOutlined,
  UpOutlined,
} from '@ant-design/icons';

const { Title, Text } = Typography;


function WiggleOptions({ isReady, isWiggling, wiggleDuration, waitTime, setWiggleDuration, setWaitTime }) {
  const [form] = Form.useForm();

  const onSubmit = async ({wiggleSpeed, waitTime}) => {
    // Wait for ready signal
    if (!isReady) return;

    // Send config to backend
    await window.backend.Wiggler.SetConfig(wiggleSpeed, waitTime).then(() => {
      message.success("Updated Wiggle Config!");
    }).catch(err => {
      message.error("Error Setting Wiggle Config");
      console.error(err);
    });

    // Restart (if already started)
    if (isWiggling) {
      await window.backend.Wiggler.StartWiggle();
    }

    // Get config from backend
    const newWiggleSpeed = await window.backend.Wiggler.GetMoveSpeed();
    const newWaitTime = await window.backend.Wiggler.GetWaitTime();

    form.setFieldsValue({
      wiggleSpeed: newWiggleSpeed,
      waitTime: newWaitTime
    });

    // Set the state data too
    setWiggleDuration(newWiggleSpeed);
    setWaitTime(newWaitTime);

  };
  const onReset = () => {
    form.resetFields();
  };


  return (
    <div className="wiggle-options">
      <Divider />

      <Title level={4}>
        Wiggler Options
      </Title>
      <Form
        name="wiggle-options"
        form={form}
        onFinish={onSubmit}
        initialValues={{
          wiggleSpeed: wiggleDuration,
          waitTime,
        }}
      >
        <Form.Item 
          label="How slowly does the mouse move?"
          name="wiggleSpeed"
        >
          <Slider 
            min={0.1}
            max={10}
            step={0.1}
            tipFormatter={(value) => `${value}s`}
            tooltipVisible
          />
        </Form.Item>
        <div style={{ height: "10px", }} />
        <Form.Item 
          label="How long does it wait between moves?"
          name="waitTime"
        >
          <Slider 
            min={0}
            max={60}
            step={1}
            tipFormatter={(value) => `${value}s`}
            tooltipVisible
          />
        </Form.Item>
        <Form.Item>
          <Space>
            <Button 
              type="primary"
              htmlType="submit"
              disabled={!isReady}
            >
              Update { isWiggling && "and Restart" }
            </Button>
            <Button 
              type="default" 
              disabled={!isReady}
              onClick={onReset}
            >
              Reset
            </Button>
          </Space>
        </Form.Item>
      </Form>
    </div>
  );
}


function App() {
  const [isReady, setIsReady] = useState(true);
  const [optionsVisible, setOptionsVisible] = useState(false);

  const [isWiggling, setIsWiggling] = useState(false);
  const [wiggleDuration, setWiggleDuration] = useState(1);
  const [waitTime, setWaitTime] = useState(2);

  // Register Wails event trigger functions
  useEffect(() => {
    // // Run when the backend is ready
    // window.backend.Wiggler.IsReady().then(() => {
    //   setIsReady(true);
    // });

    // Run when wiggler config is confirmed
    window.wails.Events.On("config-set", ({duration, time}) => {
      console.log("config-set", duration, time);
      // setWiggleDuration(duration);
      // setWaitTime(time);
    });

    // Run when a wiggler START event is confirmed
    window.wails.Events.On("wiggle-started", () => {
      setIsWiggling(true);
    });

    // Run when a wiggler STOP event is confirmed
    window.wails.Events.On("wiggle-stopped", () => {
      setIsWiggling(true);
    });

    // Run when the backend is stopping
    window.wails.Events.On("stopped", () => {
      setIsReady(false);
      setIsWiggling(false);
    });
  }, []);

  return (
    <div 
      id="app" 
      className="App"
      style={{
        maxWidth: "600px",
        margin: "0 auto",
        padding: "20px",
        // border: "1px solid #ccc",
      }}
      onKeyPress={(e) => {
        console.log(e.key);
      }}
    >
      <Title>The Wiggler</Title>
      <Space direction="vertical">
        <Text>
          Click the button to start the wiggler. Click it again to stop.
        </Text>
        <Text>
          To configure the wiggler, click the "More Options" button.
        </Text>
      </Space>

      <Divider />

      <Space direction="horizontal" size="large">
        <Space>
          <Button
            type={isWiggling ? "ghost" : "primary"}
            disabled={!isReady}
            onClick={async () => {
              if (isWiggling) {
                await window.backend.Wiggler.StopWiggle();
              } else {
                await window.backend.Wiggler.StartWiggle();
              }
              const wiggleState = await window.backend.Wiggler.IsWiggling();
              setIsWiggling(wiggleState);
            }}
          >
            { isWiggling ? "Stop" : "Start" } Wiggling
          </Button>
          <Tooltip title="Display More Options">
            <Button
              type="secondary"
              icon={optionsVisible ? <UpOutlined /> : <DownOutlined />}
              onClick={() => setOptionsVisible(!optionsVisible)}
            />
          </Tooltip>
        </Space>
        { isWiggling && <Spin /> }
      </Space>

      {optionsVisible && (
        <WiggleOptions 
          isWiggling={isWiggling}
          wiggleDuration={wiggleDuration}
          waitTime={waitTime}
          isReady={isReady}
          setWiggleDuration={setWiggleDuration}
          setWaitTime={setWaitTime}
        /> 
      )}

    </div>
  );
}

export default App;
