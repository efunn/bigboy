import pygame
import pygame.gfxdraw
import sys
import numpy as np

class delayTester(object):

    def __init__(self):
        # initialize pygame and graphics
        pygame.init()
        self.clock = pygame.time.Clock()
        self.FRAME_RATE = 60
        self.SCREEN_WIDTH, self.SCREEN_HEIGHT = 1920, 1080
        self.screen = pygame.display.set_mode(
            (self.SCREEN_WIDTH, self.SCREEN_HEIGHT), pygame.FULLSCREEN)
        # self.SCREEN_WIDTH, self.SCREEN_HEIGHT = 800, 800
        # self.screen = pygame.display.set_mode(
        #     (self.SCREEN_WIDTH, self.SCREEN_HEIGHT))
        self.BG_COLOR = 40,40,40
        self.TICK_COLOR = 150,150,150
        self.CURSOR_COLOR = 20,200,20
        self.CURSOR_WIDTH = 0.1*self.SCREEN_WIDTH

        # constants
        self.CURSOR_DELAY_SMALL_STEP = 1 # ms
        self.CURSOR_DELAY_BIG_STEP = 5 # ms
        self.PULSE_INTERVAL = 1000 # in ms
        self.PULSE_INTERVAL = self.PULSE_INTERVAL/1000.
        self.PULSE_DURATION = 150 # in ms
        self.PULSE_DURATION = self.PULSE_DURATION/1000.


        # variables
        self.cursor_time = 0
        self.cursor_delay_time = 0
        self.cursor_delay = 0 # ms

    def check_input(self):
        for event in pygame.event.get():
            if event.type == pygame.QUIT:
                self.quit()
            elif event.type == pygame.KEYDOWN:
                if event.key == pygame.K_SPACE:
                    pass
                    # self.cursor_moving = not(self.cursor_moving)
                if event.key == pygame.K_1:
                    self.cursor_delay += self.CURSOR_DELAY_BIG_STEP
                if event.key == pygame.K_2:
                    self.cursor_delay += self.CURSOR_DELAY_SMALL_STEP
                if event.key == pygame.K_3:
                    self.cursor_delay -= self.CURSOR_DELAY_SMALL_STEP
                if event.key == pygame.K_4:
                    self.cursor_delay -= self.CURSOR_DELAY_BIG_STEP
                elif event.key == pygame.K_ESCAPE: self.quit()
            elif event.type == pygame.KEYUP:
                pass

    def run(self):
        while True:
            time_passed = self.clock.tick_busy_loop(self.FRAME_RATE)/1000.
            self.check_input()
            self.draw_background()
            self.update_cursor(time_passed)
            self.draw_cursor()
            self.draw_delay_msg()
            pygame.display.flip()

    def draw_delay_msg(self):
        draw_msg(self.screen, str(self.cursor_delay)+' ms', pos=(.5*self.SCREEN_WIDTH,.5*self.SCREEN_HEIGHT))

    def update_cursor(self, time_passed):
        self.cursor_time += time_passed
        self.cursor_delay_time = self.cursor_time + self.cursor_delay/1000.
        self.cursor_time = np.mod(self.cursor_time,self.PULSE_INTERVAL)
        self.cursor_delay_time = np.mod(self.cursor_delay_time,self.PULSE_INTERVAL)
        self.cursor_fade = self.cursor_fade_function(self.cursor_time)
        self.cursor_delay_fade = self.cursor_fade_function(self.cursor_delay_time)

    def cursor_fade_function(self, cursor_time):
        if cursor_time < 0.5*self.PULSE_DURATION:
            cursor_fade = cursor_time/float(0.5*self.PULSE_DURATION)
        elif cursor_time > self.PULSE_DURATION:
            cursor_fade = 0
        else:
            cursor_fade = 1-(cursor_time-0.5*self.PULSE_DURATION)/float(0.5*self.PULSE_DURATION)
        return cursor_fade

    def draw_cursor(self):
        left_color = fade_color(self.BG_COLOR,self.CURSOR_COLOR,self.cursor_fade)
        right_color = fade_color(self.BG_COLOR,self.CURSOR_COLOR,self.cursor_delay_fade)
        draw_left_rect(self.screen, self.CURSOR_WIDTH, 0.5*self.SCREEN_HEIGHT,
            left_color, 0, 0.5*self.SCREEN_HEIGHT)
        draw_right_rect(self.screen, self.CURSOR_WIDTH, 0.5*self.SCREEN_HEIGHT,
            right_color, self.SCREEN_WIDTH, 0.5*self.SCREEN_HEIGHT)

    def draw_background(self):
        self.screen.fill(self.BG_COLOR)

    def quit(self):
        sys.exit()

def fade_color(color_1, color_2, fade_ratio):
    fade_ratio = np.clip(fade_ratio,0,1)
    return np.average([np.array(color_1),np.array(color_2)],0,[1-fade_ratio,fade_ratio]) 

def draw_left_rect(screen, width, half_height, color, x, y):
    x_coords = [x+width, x+width, x, x]
    y_coords = [y-half_height, y+half_height, y+half_height, y-half_height]
    draw_filled_aapoly(screen, list(zip(x_coords, y_coords)), color)

def draw_right_rect(screen, width, half_height, color, x, y):
    x_coords = [x, x, x-width, x-width]
    y_coords = [y-half_height, y+half_height, y+half_height, y-half_height]
    draw_filled_aapoly(screen, list(zip(x_coords, y_coords)), color)

def draw_filled_aapoly(screen, coords, color):
    pygame.gfxdraw.filled_polygon(screen, coords, color)
    pygame.gfxdraw.aapolygon(screen, coords, color)

def draw_msg(screen, text, color=(255,255,255),
             loc='center', pos=(1024/2,768/2), size=100,
             font='freesansbold.ttf'):
    font = pygame.font.Font(font, size)
    text_surf, text_rect = make_text(text, font, color)
    if loc == 'center':
        text_rect.center = pos
    elif loc == 'left':
        text_rect.center = pos
        text_rect.left = pos[0]
    elif loc == 'right':
        text_rect.center = pos
        text_rect.right = pos[0]
    screen.blit(text_surf, text_rect)

def make_text(text, font, color):
    text_surface = font.render(text, True, color)
    return text_surface, text_surface.get_rect()

if __name__ == "__main__":
    game = delayTester()
    game.run()
